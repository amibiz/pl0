package linker

import (
	"debug/macho"
	"errors"
	"fmt"
	"strings"
)

type machoFile struct {
	macho *macho.File
}

func openMacho(name string) (*machoFile, error) {
	f, err := macho.Open(name)
	if err != nil {
		return nil, err
	}
	if f.Cpu != macho.Cpu386 {
		return nil, fmt.Errorf("unsopported cpu type: %s", f.Cpu)
	}
	return &machoFile{f}, nil
}

func (f *machoFile) section(name string) (addr uint32, content []byte, err error) {
	sect := f.macho.Section(name)
	if sect == nil {
		name = strings.TrimSuffix(name, "__")
		return 0, nil, fmt.Errorf("%s section not found", name)
	}
	addr = uint32(sect.Addr)
	content, err = sect.Data()
	return
}

func (f *machoFile) symbols() ([]macho.Symbol, error) {
	var syms []macho.Symbol
	for _, s := range f.macho.Symtab.Syms {
		if s.Sect == 0 {
			return nil, fmt.Errorf("undefined symbol: %s", s.Name)
		}
		syms = append(syms)
	}
	return syms, nil
}

func (f *machoFile) entry() (uint32, error) {
	for _, s := range f.macho.Symtab.Syms {
		if s.Name == "start" && s.Sect != 0 && f.macho.Sections[s.Sect-1].Name == "__text" {
			return uint32(s.Value), nil
		}
	}
	return 0, errors.New("entry symbol not found")
}

func (f *machoFile) relocs(name string) ([]macho.Reloc, error) {
	sect := f.macho.Section(name)
	if sect == nil {
		name = strings.TrimSuffix(name, "__")
		return nil, fmt.Errorf("%s section not found", name)
	}
	return sect.Relocs, nil
}
