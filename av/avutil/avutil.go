package avutil

import (
	"bytes"
	"context"
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

// WriteHeader implements av.Muxer.
func (handler *HandlerMuxer) WriteHeader(ctx context.Context, streams []av.Stream) error {
	if handler.stage == 0 {
		if err := handler.Muxer.WriteHeader(ctx, streams); err != nil {
			return err
		}

		handler.stage++
	}

	return nil
}

// WriteTrailer implements av.Muxer.
func (handler *HandlerMuxer) WriteTrailer(ctx context.Context) error {
	if handler.stage == 1 {
		handler.stage++
		if err := handler.Muxer.WriteTrailer(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (handler *HandlerMuxer) Close() error {
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

func (handlers *Handlers) openURL(u *url.URL, uri string) (io.ReadCloser, error) { //nolint:unparam
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

func (handlers *Handlers) Open(uri string) (av.DemuxCloser, error) {
	listen := false

	if strings.HasPrefix(uri, "listen:") {
		uri = uri[len("listen:"):]
		listen = true
	}
	for _, handler := range handlers.handlers {
		if listen {
			if handler.ServerDemuxer == nil {
				continue
			}

			ok, demuxer, err := handler.ServerDemuxer(uri)
			if !ok {
				continue
			}

			return demuxer, err
		}

		if handler.URLDemuxer == nil {
			continue
		}

		if handler.URLDemuxer != nil {
			ok, demuxer, err := handler.URLDemuxer(uri)
			if !ok {
				continue
			}

			return demuxer, err
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
	if ext == "" {
		return nil, ErrOpenURLFailed
	}

	for _, handler := range handlers.handlers {
		if handler.Ext == ext {
			if handler.ReaderDemuxer != nil {
				if _, err := handlers.openURL(u, uri); err != nil {
					return nil, err
				}

				demuxer := &HandlerDemuxer{
					Demuxer: handler.ReaderDemuxer(r),
					r:       r,
				}

				return demuxer, nil
			}
		}
	}
	var probebuf [1024]byte

	if _, err := handlers.openURL(u, uri); err != nil {
		return nil, err
	}

	if _, err := io.ReadFull(r, probebuf[:]); err != nil {
		return nil, err
	}

	for _, handler := range handlers.handlers {
		if handler.Probe != nil && handler.Probe(probebuf[:]) && handler.ReaderDemuxer != nil {
			var _r io.Reader

			if rs, ok := r.(io.ReadSeeker); ok {
				if _, err := rs.Seek(0, 0); err != nil {
					return nil, err
				}

				_r = rs
			} else {
				_r = io.MultiReader(bytes.NewReader(probebuf[:]), r)
			}

			demuxer := &HandlerDemuxer{
				Demuxer: handler.ReaderDemuxer(_r),
				r:       r,
			}

			return demuxer, nil
		}
	}

	// r.Close()

	return nil, ErrOpenURLFailed
}

func (handlers *Handlers) Create(uri string) (av.MuxCloser, error) {
	_, muxer, err := handlers.FindCreate(uri)

	return muxer, err
}

func (handlers *Handlers) FindCreate(uri string) (RegisterHandler, av.MuxCloser, error) {
	listen := false

	if strings.HasPrefix(uri, "listen:") {
		uri = uri[len("listen:"):]
		listen = true
	}

	for _, handler := range handlers.handlers {
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
		return RegisterHandler{}, nil, ErrCreateMuxerFailed
	}

	for _, handler := range handlers.handlers {
		if handler.Ext == ext && handler.WriterMuxer != nil {
			w, err := handlers.createURL(uri)
			if err != nil {
				return handler, nil, err
			}

			muxer := &HandlerMuxer{
				Muxer: handler.WriterMuxer(w),
				w:     w,
			}

			return handler, muxer, err
		}
	}

	return RegisterHandler{}, nil, ErrCreateMuxerFailed
}

var DefaultHandlers = &Handlers{} //nolint:gochecknoglobals

func Open(url string) (av.DemuxCloser, error) {
	return DefaultHandlers.Open(url)
}

func Create(url string) (av.MuxCloser, error) {
	return DefaultHandlers.Create(url)
}

func CopyPackets(ctx context.Context, dst av.PacketWriter, src av.PacketReader) error {
	cc, dstCanChangeCodec := dst.(av.CodecChanger)

	for {
		pkt, err := src.ReadPacket(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return err
		}

		if pkt.NewCodecs != nil && dstCanChangeCodec {
			if err := cc.WriteCodecChange(ctx, pkt.NewCodecs); err != nil {
				return err
			}
		}

		if len(pkt.Data) == 0 {
			continue // pure codec-change notification; no media payload to write
		}

		if err := dst.WritePacket(ctx, pkt); err != nil {
			return err
		}
	}

	return nil
}

func CopyFile(ctx context.Context, dst av.Muxer, src av.Demuxer) error {
	streams, err := src.GetCodecs(ctx)
	if err != nil {
		return err
	}

	if err := dst.WriteHeader(ctx, streams); err != nil {
		return err
	}

	if err := CopyPackets(ctx, dst, src); err != nil {
		if !errors.Is(err, io.EOF) {
			return err
		}
	}

	if err := dst.WriteTrailer(ctx); err != nil {
		return err
	}

	return nil
}

func Equal(s1 []av.Stream, s2 []av.Stream) bool {
	if len(s1) != len(s2) {
		return false
	}

	for i, stream := range s1 {
		if stream.Idx != s2[i].Idx {
			return false
		}

		if stream.Codec.Type() != s2[i].Codec.Type() {
			return false
		}

		switch c1 := stream.Codec.(type) {
		case h265parser.CodecData:
			c2, ok := s2[i].Codec.(h265parser.CodecData)
			if !ok {
				return false
			}

			if eq := bytes.Compare(
				c1.AVCDecoderConfRecordBytes(),
				c2.AVCDecoderConfRecordBytes(),
			); eq != 0 {
				return false
			}
		case h264parser.CodecData:
			c2, ok := s2[i].Codec.(h264parser.CodecData)
			if !ok {
				return false
			}

			if eq := bytes.Compare(
				c1.AVCDecoderConfRecordBytes(),
				c2.AVCDecoderConfRecordBytes(),
			); eq != 0 {
				return false
			}
		case aacparser.CodecData:
			c2, ok := s2[i].Codec.(aacparser.CodecData)
			if !ok {
				return false
			}

			if eq := bytes.Compare(
				c1.MPEG4AudioConfigBytes(),
				c2.MPEG4AudioConfigBytes(),
			); eq != 0 {
				return false
			}
		}
	}

	return true
}
