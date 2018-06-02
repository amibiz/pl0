package linker

import (
	"debug/macho"
	"encoding/binary"
	"os"
)

const (
	pageSize uint32 = 4096
	pageMask        = ^(pageSize - 1)
)

// Header
const (
	CpuSubtypeX86All = 0x3
	NoUndefs         = 0x1
)

// Load command length
const (
	segment32Len  = 56
	section32Len  = 68
	unixThreadLen = 80
)

// Protection values
const (
	P_NONE  = 0x00
	P_READ  = 0x01 // read permission
	P_WRITE = 0x02 // write permission
	P_EXEC  = 0x04 // execute permission

	P_RDWR   = P_READ | P_WRITE
	P_RDEXEC = P_READ | P_EXEC
)

// Section
const (
	// Type
	S_REGULAR  = 0x0 // regular section
	S_ZEROFILL = 0x1 // zero fill on demand section

	// Attributes
	S_PURE_INSTRUCTIONS = 0x80000000 // section contains only true machine instructions
	S_SOME_INSTRUCTIONS = 0x00000400 // section contains some machine instructions
)

type unixThread struct {
	Cmd    macho.LoadCmd
	Len    uint32
	Flavor uint32
	Count  uint32
	State  [16]uint32
}

// Link creates a Mach-O executable.
func Link(dst, src string) error {
	var err error

	// Open input file
	in, err := openMacho(src)
	if err != nil {
		return err
	}

	/*
	 * Create executable layout
	 */
	var addr, ncmd, cmdsz, entry uint32

	// Segment: __PAGEZERO
	pageZero := macho.Segment32{
		Cmd:     macho.LoadCmdSegment,
		Len:     segment32Len,
		Name:    str16("__PAGEZERO"),
		Addr:    0,
		Memsz:   pageSize,
		Offset:  0,
		Filesz:  0,
		Maxprot: P_NONE,
		Prot:    P_NONE,
		Nsect:   0,
		Flag:    0,
	}
	addr = pageZero.Addr + pageZero.Memsz
	ncmd += 1
	cmdsz += pageZero.Len

	// Segment: __TEXT
	_, textContent, err := in.section("__text")
	if err != nil {
		return err
	}
	textLen := uint32(len(textContent))
	textSize := (textLen + pageSize - 1) & pageMask
	textSeg := macho.Segment32{
		Cmd:     macho.LoadCmdSegment,
		Len:     segment32Len + section32Len,
		Name:    str16("__TEXT"),
		Addr:    addr,
		Memsz:   textSize,
		Offset:  0,
		Filesz:  textSize,
		Maxprot: P_RDEXEC,
		Prot:    P_RDEXEC,
		Nsect:   1,
		Flag:    0,
	}
	textSect := macho.Section32{
		Name:     str16("__text"),
		Seg:      str16("__TEXT"),
		Addr:     textSeg.Addr + textSeg.Memsz - textLen,
		Size:     textLen,
		Offset:   textSeg.Filesz - textLen,
		Align:    0,
		Reloff:   0,
		Nreloc:   0,
		Flags:    S_PURE_INSTRUCTIONS | S_SOME_INSTRUCTIONS,
		Reserve1: 0,
		Reserve2: 0,
	}
	addr = textSeg.Addr + textSeg.Memsz
	ncmd += 1
	cmdsz += textSeg.Len

	// Segment: __DATA
	dataStart, dataContent, _ := in.section("__data")
	dataLen := uint32(len(dataContent))
	dataSize := (dataLen + pageSize - 1) & pageMask
	dataSeg := macho.Segment32{
		Cmd:     macho.LoadCmdSegment,
		Len:     segment32Len + section32Len*2,
		Name:    str16("__DATA"),
		Addr:    addr,
		Memsz:   dataSize,
		Offset:  textSeg.Offset + textSeg.Filesz,
		Filesz:  dataSize,
		Maxprot: P_RDWR,
		Prot:    P_RDWR,
		Nsect:   2,
		Flag:    0,
	}
	dataSect := macho.Section32{
		Name:     str16("__data"),
		Seg:      str16("__DATA"),
		Addr:     dataSeg.Addr,
		Size:     dataLen,
		Offset:   dataSeg.Offset,
		Align:    0,
		Reloff:   0,
		Nreloc:   0,
		Flags:    S_REGULAR,
		Reserve1: 0,
		Reserve2: 0,
	}
	_, bssContent, _ := in.section("__bss")
	bssLen := uint32(len(bssContent))
	bssSect := macho.Section32{
		Name:     str16("__bss"),
		Seg:      str16("__DATA"),
		Addr:     dataSeg.Addr,
		Size:     bssLen,
		Offset:   0,
		Align:    0,
		Reloff:   0,
		Nreloc:   0,
		Flags:    S_ZEROFILL,
		Reserve1: 0,
		Reserve2: 0,
	}
	addr = dataSeg.Addr + dataSeg.Memsz
	ncmd += 1
	cmdsz += dataSeg.Len

	// Relocate text symbols now that we got data layout
	relocs, err := in.relocs("__text")
	if err != nil {
		return err
	}
	for _, r := range relocs {
		content := textContent[r.Addr : r.Addr+1<<r.Len]
		value := binary.LittleEndian.Uint32(content)
		binary.LittleEndian.PutUint32(content, dataSeg.Addr+(value-dataStart))
	}

	// Lookup entry address
	if entry, err = in.entry(); err != nil {
		return err
	}

	// Unix Thread Command
	thread := unixThread{
		Cmd:    macho.LoadCmdUnixThread,
		Len:    unixThreadLen,
		Flavor: 0x1, // i386_THREAD_STATE
		Count:  16,
		State: [16]uint32{
			0, // AX
			0, // BX
			0, // CX
			0, // DX
			0, // DI
			0, // SI
			0, // BP
			0, // SP
			0, // SS
			0, // FLAGS
			textSect.Addr + entry, // IP
			0, // CS
			0, // DS
			0, // ES
			0, // FS
			0, // GS
		},
	}
	ncmd += 1
	cmdsz += thread.Len

	header := macho.FileHeader{
		Magic:  macho.Magic32,
		Cpu:    macho.Cpu386,
		SubCpu: CpuSubtypeX86All,
		Type:   macho.TypeExec,
		Ncmd:   ncmd,
		Cmdsz:  cmdsz,
		Flags:  NoUndefs,
	}

	// Write output file
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0700)
	if err != nil {
		return err
	}

	// Write header and load commands
	bw := &binaryWriter{w: out, bo: binary.LittleEndian}
	bw.write(header)
	bw.write(pageZero)
	bw.write(textSeg)
	bw.write(textSect)
	bw.write(dataSeg)
	bw.write(dataSect)
	bw.write(bssSect)
	bw.write(thread)

	// Write text, data and bss contents
	bw.writeAt(textContent, int64(textSect.Offset))
	bw.writeAt(make([]byte, dataSeg.Memsz), int64(dataSeg.Offset))
	bw.writeAt(dataContent, int64(dataSeg.Offset))

	if bw.err != nil {
		return bw.err
	}

	return nil
}

func str16(s string) [16]byte {
	var arr [16]byte
	copy(arr[:], s)
	return arr
}
