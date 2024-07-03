package main

import (
	"fmt"
	"io/fs"
	"net/netip"
	"os"
	"path/filepath"
	"sort"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/version"
	"github.com/pkg/errors"
)

const (
	CIDR    = "10.244.0.0/24"
	RUN_DIR = "/run/cni-ipam-state"
)

func main() {
	skel.PluginMainFuncs(skel.CNIFuncs{Add: cmdAdd, Del: cmdDel, Check: cmdCheck, GC: nil, Status: nil}, version.PluginSupports("1.0.0"), "Sample CNI")

}

func cmdAdd(args *skel.CmdArgs) error {
	return nil
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
			return errors.Wrap(err, "failed to walkdir")
		}

		if d.IsDir() {
			return nil
		}

		addr, err := netip.ParseAddr(path)
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
	addr := cidr.Addr()
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
	index := 0

	for i := 0; i < 2^cidr.Bits(); i++ {

		if addr.Compare(allocs[index]) != 0 {
			// pick allocatable address
			return addr, nil
		}
		index++
		addr = addr.Next()

	}
	return addr, fmt.Errorf("no allocatable addreesses")
}
