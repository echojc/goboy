package main

var a, f, b, c, d, e, h, l uint8
var sp uint16 = 0xfffe
var pc uint16 = 0x0100

var ram [0x2000]uint8
var vram [0x2000]uint8
var xram [0x7f]uint8

var interrupt uint8
var halted bool

var cycles int32

func Read(addr uint16) uint8 {
	switch {
	case addr < 0x4000:
		// rom #0
		return 0
	case addr < 0x8000:
		// rom #1-x
		return 0
	case addr < 0xa000:
		return vram[addr-0x8000]
	case addr < 0xc000:
		// ram #1-x
		return 0
	case addr < 0xe000:
		return ram[addr-0xc000]
	case addr < 0xfe00:
		// echo of internal ram
		return ram[addr-0xe000]
	case addr < 0xfea0:
		// oam
		return 0
	case addr < 0xff80:
		// io registers
		return 0
	case addr < 0xffff:
		return xram[addr-0xff80]
	default:
		return interrupt
	}
}

func Write(addr uint16, v uint8) {
	switch {
	case addr < 0x4000:
		// rom #0
	case addr < 0x8000:
		// rom #1-x
	case addr < 0xa000:
		vram[addr-0x8000] = v
	case addr < 0xc000:
		// ram #1-x
	case addr < 0xe000:
		ram[addr-0xc000] = v
	case addr < 0xfe00:
		// echo of internal ram
		ram[addr-0xe000] = v
	case addr < 0xfea0:
		// oam
	case addr < 0xff80:
		// io registers
	case addr < 0xffff:
		xram[addr-0xff80] = v
	default:
		// interrupt
	}
}

func Step() {
	opcode[Read(pc)]()
}

var opcode [0x255]func() = [0x255]func(){
	NOP, LD_BC_NN, LD_mBC_A, TODO, TODO, TODO, LD_B_N, TODO, LD_mNN_SP, TODO, LD_A_mBC, TODO, TODO, TODO, LD_C_N, TODO,
	TODO, LD_DE_NN, LD_mDE_A, TODO, TODO, TODO, LD_D_N, TODO, TODO, TODO, LD_A_mDE, TODO, TODO, TODO, LD_E_N, TODO,
	TODO, LD_HL_NN, LDI_mHL_A, TODO, TODO, TODO, LD_H_N, TODO, TODO, TODO, LDI_A_mHL, TODO, TODO, TODO, LD_L_N, TODO,
	TODO, LD_SP_NN, LDD_mHL_A, TODO, TODO, TODO, LD_mHL_N, TODO, TODO, TODO, LDD_A_mHL, TODO, TODO, TODO, LD_A_N, TODO,
	LD_B_B, LD_B_C, LD_B_D, LD_B_E, LD_B_H, LD_B_L, LD_B_mHL, LD_B_A, LD_C_B, LD_C_C, LD_C_D, LD_C_E, LD_C_H, LD_C_L, LD_C_mHL, LD_C_A,
	LD_D_B, LD_D_C, LD_D_D, LD_D_E, LD_D_H, LD_D_L, LD_D_mHL, LD_D_A, LD_E_B, LD_E_C, LD_E_D, LD_E_E, LD_E_H, LD_E_L, LD_E_mHL, LD_E_A,
	LD_H_B, LD_H_C, LD_H_D, LD_H_E, LD_H_H, LD_H_L, LD_H_mHL, LD_H_A, LD_L_B, LD_L_C, LD_L_D, LD_L_E, LD_L_H, LD_L_L, LD_L_mHL, LD_L_A,
	LD_mHL_B, LD_mHL_C, LD_mHL_D, LD_mHL_E, LD_mHL_H, LD_mHL_L, HALT, LD_mHL_A, LD_A_B, LD_A_C, LD_A_D, LD_A_E, LD_A_H, LD_A_L, LD_A_mHL, LD_A_A,
	TODO, TODO, TODO, TODO, TODO, TODO, TODO, TODO, TODO, TODO, TODO, TODO, TODO, TODO, TODO, TODO,
	TODO, TODO, TODO, TODO, TODO, TODO, TODO, TODO, TODO, TODO, TODO, TODO, TODO, TODO, TODO, TODO,
	TODO, TODO, TODO, TODO, TODO, TODO, TODO, TODO, TODO, TODO, TODO, TODO, TODO, TODO, TODO, TODO,
	TODO, TODO, TODO, TODO, TODO, TODO, TODO, TODO, TODO, TODO, TODO, TODO, TODO, TODO, TODO, TODO,
	TODO, POP_BC, TODO, TODO, TODO, PUSH_BC, TODO, TODO, TODO, TODO, TODO, TODO, TODO, TODO, TODO, TODO,
	TODO, POP_DE, TODO, TODO, TODO, PUSH_DE, TODO, TODO, TODO, TODO, TODO, TODO, TODO, TODO, TODO, TODO,
	LDH_mN_A, POP_HL, LDH_mC_A, TODO, TODO, PUSH_HL, TODO, TODO, TODO, TODO, LD_mNN_A, TODO, TODO, TODO, TODO, TODO,
	LDH_A_mN, POP_AF, LDH_A_mC, TODO, TODO, PUSH_AF, TODO, TODO, LD_HL_SP_N, LD_SP_HL, LD_A_mNN, TODO, TODO, TODO, TODO, TODO,
}

func TODO() { panic("unknown opcode!") }

func NOP()  { cycles += 4; pc += 1 }
func HALT() { halted = true; cycles += 4; pc += 1 }

func LD_BC_NN()   { b = Read(pc + 2); c = Read(pc + 1); cycles += 12; pc += 3 }
func LD_DE_NN()   { d = Read(pc + 2); e = Read(pc + 1); cycles += 12; pc += 3 }
func LD_HL_NN()   { h = Read(pc + 2); l = Read(pc + 1); cycles += 12; pc += 3 }
func LD_SP_NN()   { sp = uint16(Read(pc+2))<<8 + uint16(Read(pc+1)); cycles += 12; pc += 3 }
func LD_SP_HL()   { sp = uint16(h)<<8 + uint16(l); cycles += 8; pc += 1 }
func LD_HL_SP_N() { /* TODO */ }

func LD_mBC_A() { Write(uint16(b)<<8+uint16(c), a); cycles += 8; pc += 1 }
func LD_mDE_A() { Write(uint16(d)<<8+uint16(e), a); cycles += 8; pc += 1 }
func LDI_mHL_A() {
	hl := uint16(h)<<8 + uint16(l)
	Write(hl, a)
	shl := int32(hl) + 1
	h = uint8((shl >> 8) & 0xff)
	l = uint8(shl & 0xff)
	cycles += 8
	pc += 1
}
func LDD_mHL_A() {
	hl := uint16(h)<<8 + uint16(l)
	Write(hl, a)
	shl := int32(hl) - 1
	h = uint8((shl >> 8) & 0xff)
	l = uint8(shl & 0xff)
	cycles += 8
	pc += 1
}

func LD_A_mBC() { a = Read(uint16(b)<<8 + uint16(c)); cycles += 8; pc += 1 }
func LD_A_mDE() { a = Read(uint16(d)<<8 + uint16(e)); cycles += 8; pc += 1 }
func LDI_A_mHL() {
	hl := uint16(h)<<8 + uint16(l)
	a = Read(hl)
	shl := int32(hl) + 1
	h = uint8((shl >> 8) & 0xff)
	l = uint8(shl & 0xff)
	cycles += 8
	pc += 1
}
func LDD_A_mHL() {
	hl := uint16(h)<<8 + uint16(l)
	a = Read(hl)
	shl := int32(hl) - 1
	h = uint8((shl >> 8) & 0xff)
	l = uint8(shl & 0xff)
	cycles += 8
	pc += 1
}

func LD_B_N()   { b = Read(pc + 1); cycles += 8; pc += 2 }
func LD_C_N()   { c = Read(pc + 1); cycles += 8; pc += 2 }
func LD_D_N()   { d = Read(pc + 1); cycles += 8; pc += 2 }
func LD_E_N()   { e = Read(pc + 1); cycles += 8; pc += 2 }
func LD_H_N()   { h = Read(pc + 1); cycles += 8; pc += 2 }
func LD_L_N()   { l = Read(pc + 1); cycles += 8; pc += 2 }
func LD_mHL_N() { Write(uint16(h)<<8+uint16(l), Read(pc+1)); cycles += 12; pc += 2 }
func LD_A_N()   { a = Read(pc + 1); cycles += 8; pc += 2 }

func LD_B_B()   { b = b; cycles += 4; pc += 1 }
func LD_B_C()   { b = c; cycles += 4; pc += 1 }
func LD_B_D()   { b = d; cycles += 4; pc += 1 }
func LD_B_E()   { b = e; cycles += 4; pc += 1 }
func LD_B_H()   { b = h; cycles += 4; pc += 1 }
func LD_B_L()   { b = l; cycles += 4; pc += 1 }
func LD_B_mHL() { b = Read(uint16(h)<<8 + uint16(l)); cycles += 8; pc += 1 }
func LD_B_A()   { b = a; cycles += 4; pc += 1 }
func LD_C_B()   { c = b; cycles += 4; pc += 1 }
func LD_C_C()   { c = c; cycles += 4; pc += 1 }
func LD_C_D()   { c = d; cycles += 4; pc += 1 }
func LD_C_E()   { c = e; cycles += 4; pc += 1 }
func LD_C_H()   { c = h; cycles += 4; pc += 1 }
func LD_C_L()   { c = l; cycles += 4; pc += 1 }
func LD_C_mHL() { c = Read(uint16(h)<<8 + uint16(l)); cycles += 8; pc += 1 }
func LD_C_A()   { c = a; cycles += 4; pc += 1 }
func LD_D_B()   { d = b; cycles += 4; pc += 1 }
func LD_D_C()   { d = c; cycles += 4; pc += 1 }
func LD_D_D()   { d = d; cycles += 4; pc += 1 }
func LD_D_E()   { d = e; cycles += 4; pc += 1 }
func LD_D_H()   { d = h; cycles += 4; pc += 1 }
func LD_D_L()   { d = l; cycles += 4; pc += 1 }
func LD_D_mHL() { d = Read(uint16(h)<<8 + uint16(l)); cycles += 8; pc += 1 }
func LD_D_A()   { d = a; cycles += 4; pc += 1 }
func LD_E_B()   { e = b; cycles += 4; pc += 1 }
func LD_E_C()   { e = c; cycles += 4; pc += 1 }
func LD_E_D()   { e = d; cycles += 4; pc += 1 }
func LD_E_E()   { e = e; cycles += 4; pc += 1 }
func LD_E_H()   { e = h; cycles += 4; pc += 1 }
func LD_E_L()   { e = l; cycles += 4; pc += 1 }
func LD_E_mHL() { e = Read(uint16(h)<<8 + uint16(l)); cycles += 8; pc += 1 }
func LD_E_A()   { e = a; cycles += 4; pc += 1 }
func LD_H_B()   { h = b; cycles += 4; pc += 1 }
func LD_H_C()   { h = c; cycles += 4; pc += 1 }
func LD_H_D()   { h = d; cycles += 4; pc += 1 }
func LD_H_E()   { h = e; cycles += 4; pc += 1 }
func LD_H_H()   { h = h; cycles += 4; pc += 1 }
func LD_H_L()   { h = l; cycles += 4; pc += 1 }
func LD_H_mHL() { h = Read(uint16(h)<<8 + uint16(l)); cycles += 8; pc += 1 }
func LD_H_A()   { h = a; cycles += 4; pc += 1 }
func LD_L_B()   { l = b; cycles += 4; pc += 1 }
func LD_L_C()   { l = c; cycles += 4; pc += 1 }
func LD_L_D()   { l = d; cycles += 4; pc += 1 }
func LD_L_E()   { l = e; cycles += 4; pc += 1 }
func LD_L_H()   { l = h; cycles += 4; pc += 1 }
func LD_L_L()   { l = l; cycles += 4; pc += 1 }
func LD_L_mHL() { l = Read(uint16(h)<<8 + uint16(l)); cycles += 8; pc += 1 }
func LD_L_A()   { l = a; cycles += 4; pc += 1 }
func LD_mHL_B() { Write(uint16(h)<<8+uint16(l), b); cycles += 8; pc += 1 }
func LD_mHL_C() { Write(uint16(h)<<8+uint16(l), c); cycles += 8; pc += 1 }
func LD_mHL_D() { Write(uint16(h)<<8+uint16(l), d); cycles += 8; pc += 1 }
func LD_mHL_E() { Write(uint16(h)<<8+uint16(l), e); cycles += 8; pc += 1 }
func LD_mHL_H() { Write(uint16(h)<<8+uint16(l), h); cycles += 8; pc += 1 }
func LD_mHL_L() { Write(uint16(h)<<8+uint16(l), l); cycles += 8; pc += 1 }
func LD_mHL_A() { Write(uint16(h)<<8+uint16(l), a); cycles += 8; pc += 1 }
func LD_A_B()   { a = b; cycles += 4; pc += 1 }
func LD_A_C()   { a = c; cycles += 4; pc += 1 }
func LD_A_D()   { a = d; cycles += 4; pc += 1 }
func LD_A_E()   { a = e; cycles += 4; pc += 1 }
func LD_A_H()   { a = h; cycles += 4; pc += 1 }
func LD_A_L()   { a = l; cycles += 4; pc += 1 }
func LD_A_mHL() { a = Read(uint16(h)<<8 + uint16(l)); cycles += 8; pc += 1 }
func LD_A_A()   { a = a; cycles += 4; pc += 1 }

func PUSH_BC() { Write(sp-1, b); Write(sp-2, c); sp -= 2; cycles += 16; pc += 1 }
func PUSH_DE() { Write(sp-1, d); Write(sp-2, e); sp -= 2; cycles += 16; pc += 1 }
func PUSH_HL() { Write(sp-1, h); Write(sp-2, l); sp -= 2; cycles += 16; pc += 1 }
func PUSH_AF() { Write(sp-1, a); Write(sp-2, f); sp -= 2; cycles += 16; pc += 1 }

func POP_BC() { c = Read(sp); b = Read(sp + 1); sp += 2; cycles += 12; pc += 1 }
func POP_DE() { e = Read(sp); d = Read(sp + 1); sp += 2; cycles += 12; pc += 1 }
func POP_HL() { l = Read(sp); h = Read(sp + 1); sp += 2; cycles += 12; pc += 1 }
func POP_AF() { f = Read(sp); a = Read(sp + 1); sp += 2; cycles += 12; pc += 1 }

func LDH_mN_A() { Write(0xff00+uint16(Read(pc+1)), a); cycles += 12; pc += 2 }
func LDH_A_mN() { a = Read(0xff00 + uint16(Read(pc+1))); cycles += 12; pc += 2 }
func LDH_mC_A() { Write(0xff00+uint16(c), a); cycles += 8; pc += 1 }
func LDH_A_mC() { a = Read(0xff00 + uint16(c)); cycles += 8; pc += 1 }

func LD_mNN_A() { Write(uint16(Read(pc+2))<<8+uint16(Read(pc+1)), a); cycles += 16; pc += 3 }
func LD_A_mNN() { a = Read(uint16(Read(pc+2))<<8 + uint16(Read(pc+1))); cycles += 16; pc += 3 }
func LD_mNN_SP() {
	m := uint16(Read(pc+2))<<8 + uint16(Read(pc+1))
	Write(m, uint8(sp&0xff))
	Write(m+1, uint8((sp>>8)&0xff))
	cycles += 20
	pc += 3
}
