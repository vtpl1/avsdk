package streammanager4_test

import (
	"testing"

	"github.com/vtpl1/avsdk/av"
	"github.com/vtpl1/avsdk/av/streammanager4"
)

func TestInterfaceImplementations(t *testing.T) {
	var _ av.StreamManager = (*streammanager4.StreamManager)(nil)
	var _ av.Stopper = (*streammanager4.StreamManager)(nil)
	var _ av.Stopper = (*streammanager4.Producer)(nil)
}
