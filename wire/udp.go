/*
* @Author: BlahGeek
* @Date:   2015-06-24
* @Last Modified by:   BlahGeek
* @Last Modified time: 2015-08-18
 */

package wire

import "net"
import "fmt"
import log "github.com/Sirupsen/logrus"
import "encoding/json"

const UDP_DEFAULT_MTU = 1450

type UDPTransportOptions struct {
	ServerAddr string  `json:"server_addr"`
	ClientAddr string  `json:"client_addr"`
	MTU        float64 `json:"mtu"`
}

type UDPTransport struct {
	udp         *net.UDPConn
	remote_addr *net.UDPAddr
	is_server   bool
	mtu         int

	logger *log.Entry
}

func (trans *UDPTransport) String() string {
	return fmt.Sprintf("UDP[%v]", trans.remote_addr)
}

func (trans *UDPTransport) MTU() int {
	return trans.mtu
}

func (trans *UDPTransport) Open(is_server bool, options json.RawMessage) error {
	var server_addr, client_addr *net.UDPAddr
	var err error

	trans.logger = log.WithField("logger", "UDPTransport")

	var opt UDPTransportOptions
	if err = json.Unmarshal(options, &opt); err != nil {
		return err
	}

	trans.mtu = UDP_DEFAULT_MTU
	if opt.MTU > 0 {
		trans.mtu = int(opt.MTU)
	}

	trans.is_server = is_server

	if server_addr, err = net.ResolveUDPAddr("udp", opt.ServerAddr); err != nil {
		return fmt.Errorf("Error resolving server addr: %v", err)
	}
	if len(opt.ClientAddr) > 0 {
		if client_addr, err = net.ResolveUDPAddr("udp", opt.ClientAddr); err != nil {
			return fmt.Errorf("Error resolving client addr: %v", err)
		}
	}

	if is_server {
		trans.logger.WithField("addr", server_addr).Info("Listening on address")
		trans.udp, err = net.ListenUDP("udp", server_addr)
		if err != nil {
			return fmt.Errorf("Error listening UDP: %v", err)
		}
	} else {
		trans.logger.WithFields(log.Fields{
			"server": server_addr,
			"local":  client_addr,
		}).Info("Dialing to address")
		trans.udp, err = net.DialUDP("udp", client_addr, server_addr)
		if err != nil {
			return fmt.Errorf("Error dialing UDP: %v", err)
		}
		trans.remote_addr = server_addr
	}

	return nil
}

func (trans *UDPTransport) GetWireNetworks() []net.IPNet {
	if trans.is_server {
		return make([]net.IPNet, 0)
	}
	mask_len := len(trans.remote_addr.IP) * 8
	return []net.IPNet{
		net.IPNet{trans.remote_addr.IP, net.CIDRMask(mask_len, mask_len)},
	}
}

func (trans *UDPTransport) Close() error {
	if trans.udp == nil {
		return nil
	}
	return trans.udp.Close()
}

func (trans *UDPTransport) Read(buf []byte) (int, error) {
	rdlen, addr, err := trans.udp.ReadFromUDP(buf)
	if trans.is_server && err == nil {
		trans.remote_addr = addr
	}

	return rdlen, err
}

func (trans *UDPTransport) Write(buf []byte) (int, error) {
	if trans.remote_addr == nil && trans.is_server {
		return 0, nil
	}

	if trans.is_server {
		return trans.udp.WriteToUDP(buf, trans.remote_addr)
	}
	return trans.udp.Write(buf)
}
