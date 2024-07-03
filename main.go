package main

import (
	"io/fs"
	"net"
	"net/netip"
	"os"
	"path/filepath"
	"sort"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/cni/pkg/version"
	"github.com/pkg/errors"
)

const (
	CIDR        = "10.244.0.0/24"
	RUN_DIR     = "/run/cni-ipam-state"
	CNI_VERSION = "1.0.0"
)

func main() {
	skel.PluginMainFuncs(skel.CNIFuncs{Add: cmdAdd, Del: cmdDel, Check: cmdCheck, GC: nil, Status: nil}, version.PluginSupports("1.0.0"), "Sample CNI")

}

func cmdAdd(args *skel.CmdArgs) error {
	addr, err := pick()
	if err != nil {
		return err
	}
	if err := createAllocation(addr); err != nil {
		return err
	}

	cidr, err := netip.ParsePrefix(CIDR)
	if err != nil {
		return err
	}

	iface := 0
	a := &net.IPNet{
		IP:   net.ParseIP(addr.String()),
		Mask: net.CIDRMask(24, 32),
	}

	_, dst, err := net.ParseCIDR("0.0.0.0/0")
	if err != nil {
		return err
	}

	gw := net.ParseIP(cidr.Addr().Next().String())

	result := current.Result{
		CNIVersion: CNI_VERSION,
		Interfaces: []*current.Interface{
			{
				Name: "eth0",
			},
		},
		IPs: []*current.IPConfig{
			{
				Interface: &iface,
				Gateway:   gw,
				Address:   *a,
			},
		},
		Routes: []*types.Route{
			{
				Dst: *dst,
				GW:  gw,
			},
		},
	}
	return types.PrintResult(&result, CNI_VERSION)
}

func cmdDel(args *skel.CmdArgs) error {
	return nil
}

func cmdCheck(args *skel.CmdArgs) error {
	return nil
}

func listAllocation() ([]netip.Addr, error) {
	allocations := make([]netip.Addr, 0)
	if err := filepath.WalkDir(RUN_DIR, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				os.MkdirAll(RUN_DIR, os.ModePerm)
			} else {
				return errors.Wrap(err, "failed to walkdir")
			}
		}

		if d.IsDir() {
			return nil
		}

		addr, err := netip.ParseAddr(filepath.Base(path))
		if err != nil {
			return err
		}

		allocations = append(allocations, addr)

		return nil
	}); err != nil {
		return nil, err
	}
	sort.Slice(allocations, func(i, j int) bool {
		return allocations[i].Compare(allocations[j]) < 0
	})
	return allocations, nil
}

func createAllocation(addr netip.Addr) error {
	file, err := os.Create(filepath.Join(RUN_DIR, addr.String()))
	if err != nil {
		return err
	}
	defer file.Close()
	return nil
}

func deleteAllocation(addr netip.Addr) error {
	return os.Remove(filepath.Join(RUN_DIR, addr.String()))
}

func pick() (netip.Addr, error) {
	cidr, err := netip.ParsePrefix(CIDR)
	addr := cidr.Addr().Next().Next() // .2
	if err != nil {
		return addr, err
	}

	allocs, err := listAllocation()
	if err != nil {
		return addr, err
	}

	if len(allocs) == 0 {
		return addr, nil
	}

	max := allocs[len(allocs)-1]

	addr = max.Next()

	return addr, nil
}
