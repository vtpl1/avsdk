package avutil

import (
	"bytes"
	"errors"
	"io"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/vtpl1/avsdk/av"
	"github.com/vtpl1/avsdk/codec/aacparser"
	"github.com/vtpl1/avsdk/codec/h264parser"
	"github.com/vtpl1/avsdk/codec/h265parser"
)

var (
	ErrOpenURLFailed     = errors.New("openUrl failed")
	ErrCreateMuxerFailed = errors.New("create muxer failed")
)

type HandlerDemuxer struct {
	av.Demuxer
	r io.ReadCloser
}

func (handler *HandlerDemuxer) Close() error {
	return handler.r.Close()
}

type HandlerMuxer struct {
	av.Muxer
	w     io.WriteCloser
	stage int
}

func (handler *HandlerMuxer) WriteHeader(streams []av.CodecData) error {
	if handler.stage == 0 {
		if err := handler.Muxer.WriteHeader(streams); err != nil {
			return err
		}

		handler.stage++
	}

	return nil
}

func (handler *HandlerMuxer) WriteTrailer() error {
	if handler.stage == 1 {
		handler.stage++
		if err := handler.Muxer.WriteTrailer(); err != nil {
			return err
		}
	}

	return nil
}

func (handler *HandlerMuxer) Close() error {
	if err := handler.WriteTrailer(); err != nil {
		return err
	}

	return handler.w.Close()
}

type RegisterHandler struct {
	Ext           string
	ReaderDemuxer func(io.Reader) av.Demuxer
	WriterMuxer   func(io.Writer) av.Muxer
	URLMuxer      func(string) (bool, av.MuxCloser, error)
	URLDemuxer    func(string) (bool, av.DemuxCloser, error)
	URLReader     func(string) (bool, io.ReadCloser, error)
	Probe         func([]byte) bool
	// AudioEncoder  func(av.CodecType) (av.AudioEncoder, error)
	// AudioDecoder  func(av.AudioCodecData) (av.AudioDecoder, error)
	ServerDemuxer func(string) (bool, av.DemuxCloser, error)
	ServerMuxer   func(string) (bool, av.MuxCloser, error)
	CodecTypes    []av.CodecType
}

type Handlers struct {
	handlers []RegisterHandler
}

func (handlers *Handlers) Add(fn func(*RegisterHandler)) {
	handler := &RegisterHandler{}
	fn(handler)
	handlers.handlers = append(handlers.handlers, *handler)
}

func (handlers *Handlers) openURL(u *url.URL, uri string) (io.ReadCloser, error) {
	if u != nil && u.Scheme != "" {
		for _, handler := range handlers.handlers {
			if handler.URLReader != nil {
				if ok, r, err := handler.URLReader(uri); ok {
					return r, err
				}
			}
		}

		return nil, ErrOpenURLFailed
	}

	return os.Open(uri)
}

func (handlers *Handlers) createURL(uri string) (io.WriteCloser, error) {
	return os.Create(uri)
}

// func (self *Handlers) NewAudioEncoder(typ av.CodecType) (enc av.AudioEncoder, err error) {
// 	for _, handler := range self.handlers {
// 		if handler.AudioEncoder != nil {
// 			if enc, _ = handler.AudioEncoder(typ); enc != nil {
// 				return
// 			}
// 		}
// 	}
// 	err = fmt.Errorf("avutil: encoder", typ, "not found")
// 	return
// }

// func (self *Handlers) NewAudioDecoder(codec av.AudioCodecData) (dec av.AudioDecoder, err error) {
// 	for _, handler := range self.handlers {
// 		if handler.AudioDecoder != nil {
// 			if dec, _ = handler.AudioDecoder(codec); dec != nil {
// 				return
// 			}
// 		}
// 	}
// 	err = fmt.Errorf("avutil: decoder", codec.Type(), "not found")
// 	return
// }

func (handlers *Handlers) Open(uri string) (demuxer av.DemuxCloser, err error) {
	listen := false

	if strings.HasPrefix(uri, "listen:") {
		uri = uri[len("listen:"):]
		listen = true
	}

	for _, handler := range handlers.handlers {
		if listen {
			if handler.ServerDemuxer != nil {
				var ok bool
				if ok, demuxer, err = handler.ServerDemuxer(uri); ok {
					return demuxer, err
				}
			}
		} else {
			if handler.URLDemuxer != nil {
				var ok bool
				if ok, demuxer, err = handler.URLDemuxer(uri); ok {
					return demuxer, err
				}
			}
		}
	}

	var r io.ReadCloser

	var ext string

	var u *url.URL
	if u, _ = url.Parse(uri); u != nil && u.Scheme != "" {
		ext = path.Ext(u.Path)
	} else {
		ext = path.Ext(uri)
	}

	if ext != "" {
		for _, handler := range handlers.handlers {
			if handler.Ext == ext {
				if handler.ReaderDemuxer != nil {
					if r, err = handlers.openURL(u, uri); err != nil {
						return demuxer, err
					}

					demuxer = &HandlerDemuxer{
						Demuxer: handler.ReaderDemuxer(r),
						r:       r,
					}

					return demuxer, err
				}
			}
		}
	}

	var probebuf [1024]byte

	if r, err = handlers.openURL(u, uri); err != nil {
		return demuxer, err
	}

	if _, err = io.ReadFull(r, probebuf[:]); err != nil {
		return demuxer, err
	}

	for _, handler := range handlers.handlers {
		if handler.Probe != nil && handler.Probe(probebuf[:]) && handler.ReaderDemuxer != nil {
			var _r io.Reader

			if rs, ok := r.(io.ReadSeeker); ok {
				if _, err = rs.Seek(0, 0); err != nil {
					return demuxer, err
				}

				_r = rs
			} else {
				_r = io.MultiReader(bytes.NewReader(probebuf[:]), r)
			}

			demuxer = &HandlerDemuxer{
				Demuxer: handler.ReaderDemuxer(_r),
				r:       r,
			}

			return demuxer, err
		}
	}

	r.Close()

	err = ErrOpenURLFailed

	return demuxer, err
}

func (handlers *Handlers) Create(uri string) (muxer av.MuxCloser, err error) {
	_, muxer, err = handlers.FindCreate(uri)

	return
}

func (handlers *Handlers) FindCreate(uri string) (handler RegisterHandler, muxer av.MuxCloser, err error) {
	listen := false

	if strings.HasPrefix(uri, "listen:") {
		uri = uri[len("listen:"):]
		listen = true
	}

	for _, handler = range handlers.handlers {
		if listen {
			if handler.ServerMuxer == nil {
				continue
			}

			ok, muxer, err := handler.ServerMuxer(uri)
			if !ok {
				continue
			}

			return handler, muxer, err
		}

		if handler.URLMuxer == nil {
			continue
		}

		if handler.URLMuxer != nil {
			ok, muxer, err := handler.URLMuxer(uri)
			if !ok {
				continue
			}

			return handler, muxer, err
		}
	}

	var ext string

	var u *url.URL
	if u, _ = url.Parse(uri); u != nil && u.Scheme != "" {
		ext = path.Ext(u.Path)
	} else {
		ext = path.Ext(uri)
	}

	if ext == "" {
		return handler, nil, ErrCreateMuxerFailed
	}

	for _, handler = range handlers.handlers {
		if handler.Ext == ext && handler.WriterMuxer != nil {
			var w io.WriteCloser

			if w, err = handlers.createURL(uri); err != nil {
				return handler, muxer, err
			}

			muxer = &HandlerMuxer{
				Muxer: handler.WriterMuxer(w),
				w:     w,
			}

			return handler, muxer, err
		}
	}

	return handler, muxer, ErrCreateMuxerFailed
}

var DefaultHandlers = &Handlers{}

func Open(url string) (demuxer av.DemuxCloser, err error) {
	return DefaultHandlers.Open(url)
}

func Create(url string) (muxer av.MuxCloser, err error) {
	return DefaultHandlers.Create(url)
}

func CopyPackets(dst av.PacketWriter, src av.PacketReader) error {
	for {
		pkt, err := src.ReadPacket()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return err
		}

		if err := dst.WritePacket(pkt); err != nil {
			return err
		}
	}

	return nil
}

func CopyFile(dst av.Muxer, src av.Demuxer) error {
	streams, err := src.Streams()
	if err != nil {
		return err
	}

	if err := dst.WriteHeader(streams); err != nil {
		return err
	}

	if err := CopyPackets(dst, src); err != nil {
		if !errors.Is(err, io.EOF) {
			return err
		}
	}

	if err := dst.WriteTrailer(); err != nil {
		return err
	}

	return nil
}

func Equal(c1 []av.CodecData, c2 []av.CodecData) bool {
	if len(c1) != len(c2) {
		return false
	}

	for i, codec := range c1 {
		if codec.Type() != c2[i].Type() {
			return false
		}

		switch c1codec := codec.(type) {
		case h265parser.CodecData:
			c2codec, ok := c2[i].(h265parser.CodecData)
			if !ok {
				return false
			}

			if eq := bytes.Compare(
				c1codec.AVCDecoderConfRecordBytes(),
				c2codec.AVCDecoderConfRecordBytes(),
			); eq != 0 {
				return false
			}
		case h264parser.CodecData:
			c2codec, ok := c2[i].(h264parser.CodecData)
			if !ok {
				return false
			}

			if eq := bytes.Compare(
				c1codec.AVCDecoderConfRecordBytes(),
				c2codec.AVCDecoderConfRecordBytes(),
			); eq != 0 {
				return false
			}
		case aacparser.CodecData:
			c2codec, ok := c2[i].(aacparser.CodecData)
			if !ok {
				return false
			}

			if eq := bytes.Compare(
				c1codec.MPEG4AudioConfigBytes(),
				c2codec.MPEG4AudioConfigBytes(),
			); eq != 0 {
				return false
			}
		}
	}

	return true
}
