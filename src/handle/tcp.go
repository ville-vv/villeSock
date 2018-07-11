package handle

import (
	"io"
	"net"
	"time"
	//"strings"
	"github.com/shadowsocks/go-shadowsocks2/socks"
	//"fmt"
	"common/villog"
)

// Create a SOCKS server listening on addr and proxy to server.
func socksLocal(addr, server string, shadow func(net.Conn) net.Conn) {
	//logf("SOCKS proxy %s <-> %s", addr, server)
	tcpLocal(addr, server, shadow, func(c net.Conn) (socks.Addr, error) { return socks.Handshake(c) })
}

// Create a TCP tunnel from addr to target via server.
func tcpTun(addr, server, target string, shadow func(net.Conn) net.Conn) {
	tgt := socks.ParseAddr(target)
	if tgt == nil {
		//logf("invalid target address %q", target)
		return
	}
	//logf("TCP tunnel %s <-> %s <-> %s", addr, server, target)
	tcpLocal(addr, server, shadow, func(net.Conn) (socks.Addr, error) { return tgt, nil })
}

// Listen on addr and proxy to server to reach target from getAddr.
func tcpLocal(addr, server string, shadow func(net.Conn) net.Conn, getAddr func(net.Conn) (socks.Addr, error)) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		//logf("failed to listen on %s: %v", addr, err)
		return
	}

	for {
		c, err := l.Accept()
		if err != nil {
			//logf("failed to accept: %s", err)
			continue
		}

		go func() {
			defer c.Close()
			c.(*net.TCPConn).SetKeepAlive(true)
			tgt, err := getAddr(c)
			if err != nil {

				// UDP: keep the connection until disconnect then free the UDP socket
				if err == socks.InfoUDPAssociate {
					buf := []byte{}
					// block here
					for {
						_, err := c.Read(buf)
						if err, ok := err.(net.Error); ok && err.Timeout() {
							continue
						}
						//logf("UDP Associate End.")
						return
					}
				}

				//logf("failed to get target address: %v", err)
				return
			}

			rc, err := net.Dial("tcp", server)
			if err != nil {
				//logf("failed to connect to server %v: %v", server, err)
				return
			}
			defer rc.Close()
			rc.(*net.TCPConn).SetKeepAlive(true)
			rc = shadow(rc)

			if _, err = rc.Write(tgt); err != nil {
				villog.LogE("failed to send target address: %v", err)
				return
			}

			villog.LogI("proxy %s <-> %s <-> %s", c.RemoteAddr(), server, tgt)
			_, _, err = relay(rc, c)
			if err != nil {
				if err, ok := err.(net.Error); ok && err.Timeout() {
					return // ignore i/o timeout
				}
				villog.LogE("relay error: %v", err)
			}
		}()
	}
}

// Listen on addr for incoming connections.
func TcpRemote(addr string, shadow func(net.Conn) net.Conn) {
	//启动服务监听
	l, err := net.Listen("tcp", addr)
	if err != nil {
		villog.LogE("failed to listen on %s: %v", addr, err)
		return
	}

	villog.LogI("listening TCP on %s", addr)
	for {
		c, err := l.Accept()
		if err != nil {
			villog.LogE("failed to accept: %v", err)
			continue
		}
		
		go func() {
			defer c.Close()
			c.(*net.TCPConn).SetKeepAlive(true)
			/**
			 * 数据解密
			 */
			c = shadow(c)

			/**
			 * 获取访问的目的地址
			 */
			tgt, err := socks.ReadAddr(c)
			if err != nil {
				villog.LogE("failed to get target address: %v", err)
				return
			}

			/**
			 * 建立要访问的目的地主 远程连接
			 */
			rc, err := net.Dial("tcp", tgt.String())
			if err != nil {
				villog.LogE("failed to connect to target: %v", err)
				return
			}
			defer rc.Close()
			rc.(*net.TCPConn).SetKeepAlive(true)

			villog.LogI("had proxy origin:%s <-> target:%s", c.RemoteAddr(), tgt)
			_, _, err = relay(c, rc)
			if err != nil {
				if err, ok := err.(net.Error); ok && err.Timeout() {
					return // ignore i/o timeout
				}
				villog.LogE("relay error: %v", err)
			}
		}()
	}
}

func dtCopy(dst io.Writer, src io.Reader, buf []byte, flag int32) (written int64, err error) {
	// If the reader has a WriteTo method, use it to do the copy.
	// Avoids an allocation and a copy.

	//if wt, ok := src.(io.WriterTo); ok {
	//	villog.LogI("WriterTo")
	//	return wt.WriteTo(dst)
	//}
	//// Similarly, if the writer has a ReadFrom method, use it to do the copy.
	//if rt, ok := dst.(io.ReaderFrom); ok {
	//	villog.LogI("ReaderFrom")
	//	return rt.ReadFrom(src)
	//}
	if buf == nil {
		buf = make([]byte, 32*1024)
	}
	for {
		villog.LogI("等待数据转发 %d", flag)
		nr, er := src.Read(buf)
		if nr > 0 {
			villog.LogI("%d data length:%d",flag,nr)
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	return written, err
}
func DtCopy(dst io.Writer, src io.Reader, flag int32) (written int64, err error) {
	return dtCopy(dst, src, nil, flag)
}

// relay copies between left and right bidirectionally. Returns number of
// bytes copied from right to left, from left to right, and any error occurred.
func relay(left, right net.Conn) (int64, int64, error) {
	type res struct {
		N   int64
		Err error
	}
	ch := make(chan res)

	go func() {
		villog.LogI("开始 target = %v -> origin = %v", right.LocalAddr(), left.LocalAddr())
		n, err := io.Copy(right, left)
		right.SetDeadline(time.Now()) // wake up the other goroutine blocking on right
		left.SetDeadline(time.Now())  // wake up the other goroutine blocking on left
		ch <- res{n, err}
	}()

	villog.LogI("后执行这个 origin = %v -> target = %v", left.LocalAddr(), right.LocalAddr())
	n, err := io.Copy(left, right)
	right.SetDeadline(time.Now()) // wake up the other goroutine blocking on right
	left.SetDeadline(time.Now())  // wake up the other goroutine blocking on left
	rs := <-ch
	villog.LogI("转发数据完成：")

	if err == nil {
		err = rs.Err
	}
	return n, 0, err
}
