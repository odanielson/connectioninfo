// Package connectioninfo implements functions to lookup information about
// network connections.
package connectioninfo

import (
    "bufio"
    "bytes"
    "fmt"
    "github.com/odanielson/pidinfo"
    "io/ioutil"
    "net"
    "strconv"
)

// Conn represents a network connection.
type Conn struct {
    src      net.IP
    src_port int
    dst      net.IP
    dst_port int
}

// ConnInfo contains meta information about a connection.
type ConnInfo struct {
    conn        Conn
    inode       int
    processInfo pidinfo.ProcessInfo
}

// LookupTcpConnection returns the inode number for connection given by a_conn
func LookupTcpConnection(a_conn Conn) int {
    if data, err := ioutil.ReadFile("/proc/net/tcp"); err == nil {
        reader := bytes.NewReader(data)
        scanner := bufio.NewScanner(reader)
        for scanner.Scan() {
            entry := parseLine(scanner.Text())
            if entry.conn.match(&a_conn) {
                var info = pidinfo.FindProcessForInode(entry.inode)
                fmt.Printf("Found in cmd %s (pid = %d)\n", info.Cmd,
                    info.Pid)
                return entry.inode
            }
        }
    }
    return -1

}

func (a *Conn) match(b *Conn) bool {
    if a.src.Equal(b.src) &&
        a.dst.Equal(b.dst) &&
        a.src_port == b.src_port &&
        a.dst_port == b.dst_port {
        return true
    }
    return false
}

func parseLine(line string) ConnInfo {
    var entry ConnInfo
    var src, src_port, dst, dst_port, index, garbage, inode int

    //                 sl  local remo  st tx_q  rq_q  re ui ti ino
    fmt.Sscanf(line, " %d: %X:%X %X:%X %X %X:%X %X:%X %X %d %d %d",
        &index, &src, &src_port, &dst, &dst_port,
        &garbage, &garbage, &garbage, &garbage,
        &garbage, &garbage, &garbage, &garbage,
        &inode)

    entry.conn.src = []byte{byte(src & 0xff), byte(src >> 8 & 0xff),
        byte(src >> 16 & 0xff), byte(src >> 24 & 0xff)}
    entry.conn.src_port = src_port

    entry.conn.dst = []byte{byte(dst & 0xff), byte(dst >> 8 & 0xff),
        byte(dst >> 16 & 0xff), byte(dst >> 24 & 0xff)}
    entry.conn.dst_port = dst_port
    entry.inode = inode
    return entry
}

// Print out a representation of a_conn
func PrintConn(a_conn Conn) {
    fmt.Printf("%s:%d -> %s:%d\n", a_conn.src, a_conn.src_port, a_conn.dst,
        a_conn.dst_port)
}

// Parse params a_src and a_dst into a Conn object
func ParseConn(a_src string, a_dst string) Conn {
    var c Conn
    var src, port string
    var err error
    if src, port, err = net.SplitHostPort(a_src); err == nil {
        c.src = net.ParseIP(src)
        c.src_port, err = strconv.Atoi(port)
    }
    if src, port, err = net.SplitHostPort(a_dst); err == nil {
        c.dst = net.ParseIP(src)
        c.dst_port, err = strconv.Atoi(port)
    }
    return c
}
