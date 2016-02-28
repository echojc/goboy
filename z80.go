package main

var a, b, c, d, e, h, l uint8
var fz, fn, fh, fc bool
var sp uint16 = 0xfffe
var pc uint16 = 0x0100

var ram [0x2000]uint8
var vram [0x2000]uint8
var xram [0x7f]uint8

var interrupt uint8
var halted bool

var cycles int32

func Step() {
	opcode[read(pc)]()
}

func read(addr uint16) uint8 {
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

func write(addr uint16, v uint8) {
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

var opcode [0x100]func() = [0x100]func(){
	NOP, LD_BC_NN, LD_mBC_A, TODO, INC_B, DEC_B, LD_B_N, TODO, LD_mNN_SP, ADD_HL_BC, LD_A_mBC, TODO, INC_C, DEC_C, LD_C_N, TODO,
	TODO, LD_DE_NN, LD_mDE_A, TODO, INC_D, DEC_D, LD_D_N, TODO, TODO, ADD_HL_DE, LD_A_mDE, TODO, INC_E, DEC_E, LD_E_N, TODO,
	TODO, LD_HL_NN, LDI_mHL_A, TODO, INC_H, DEC_H, LD_H_N, TODO, TODO, ADD_HL_HL, LDI_A_mHL, TODO, INC_L, DEC_L, LD_L_N, TODO,
	TODO, LD_SP_NN, LDD_mHL_A, TODO, INC_mHL, DEC_mHL, LD_mHL_N, TODO, TODO, ADD_HL_SP, LDD_A_mHL, TODO, INC_A, DEC_A, LD_A_N, TODO,
	LD_B_B, LD_B_C, LD_B_D, LD_B_E, LD_B_H, LD_B_L, LD_B_mHL, LD_B_A, LD_C_B, LD_C_C, LD_C_D, LD_C_E, LD_C_H, LD_C_L, LD_C_mHL, LD_C_A,
	LD_D_B, LD_D_C, LD_D_D, LD_D_E, LD_D_H, LD_D_L, LD_D_mHL, LD_D_A, LD_E_B, LD_E_C, LD_E_D, LD_E_E, LD_E_H, LD_E_L, LD_E_mHL, LD_E_A,
	LD_H_B, LD_H_C, LD_H_D, LD_H_E, LD_H_H, LD_H_L, LD_H_mHL, LD_H_A, LD_L_B, LD_L_C, LD_L_D, LD_L_E, LD_L_H, LD_L_L, LD_L_mHL, LD_L_A,
	LD_mHL_B, LD_mHL_C, LD_mHL_D, LD_mHL_E, LD_mHL_H, LD_mHL_L, HALT, LD_mHL_A, LD_A_B, LD_A_C, LD_A_D, LD_A_E, LD_A_H, LD_A_L, LD_A_mHL, LD_A_A,
	ADD_A_B, ADD_A_C, ADD_A_D, ADD_A_E, ADD_A_H, ADD_A_L, ADD_A_mHL, ADD_A_A, ADC_A_B, ADC_A_C, ADC_A_D, ADC_A_E, ADC_A_H, ADC_A_L, ADC_A_mHL, ADC_A_A,
	SUB_A_B, SUB_A_C, SUB_A_D, SUB_A_E, SUB_A_H, SUB_A_L, SUB_A_mHL, SUB_A_A, SBC_A_B, SBC_A_C, SBC_A_D, SBC_A_E, SBC_A_H, SBC_A_L, SBC_A_mHL, SBC_A_A,
	AND_B, AND_C, AND_D, AND_E, AND_H, AND_L, AND_mHL, AND_A, XOR_B, XOR_C, XOR_D, XOR_E, XOR_H, XOR_L, XOR_mHL, XOR_A,
	OR_B, OR_C, OR_D, OR_E, OR_H, OR_L, OR_mHL, OR_A, CP_B, CP_C, CP_D, CP_E, CP_H, CP_L, CP_mHL, CP_A,
	TODO, POP_BC, TODO, TODO, TODO, PUSH_BC, ADD_A_N, TODO, TODO, TODO, TODO, TODO, TODO, TODO, ADC_A_N, TODO,
	TODO, POP_DE, TODO, TODO, TODO, PUSH_DE, SUB_A_N, TODO, TODO, TODO, TODO, TODO, TODO, TODO, SBC_A_N, TODO,
	LDH_mN_A, POP_HL, LDH_mC_A, TODO, TODO, PUSH_HL, AND_N, TODO, ADD_SP_N, TODO, LD_mNN_A, TODO, TODO, TODO, XOR_N, TODO,
	LDH_A_mN, POP_AF, LDH_A_mC, TODO, TODO, PUSH_AF, OR_N, TODO, LD_HL_SP_N, LD_SP_HL, LD_A_mNN, TODO, TODO, TODO, CP_N, TODO,
}

func TODO() { panic("unknown opcode!") }

func NOP()  { cycles += 4; pc += 1 }
func HALT() { halted = true; cycles += 4; pc += 1 }

func LD_BC_NN() { b = read(pc + 2); c = read(pc + 1); cycles += 12; pc += 3 }
func LD_DE_NN() { d = read(pc + 2); e = read(pc + 1); cycles += 12; pc += 3 }
func LD_HL_NN() { h = read(pc + 2); l = read(pc + 1); cycles += 12; pc += 3 }
func LD_SP_NN() { sp = uint16(read(pc+2))<<8 + uint16(read(pc+1)); cycles += 12; pc += 3 }
func LD_SP_HL() { sp = uint16(h)<<8 + uint16(l); cycles += 8; pc += 1 }

func LD_mBC_A() { write(uint16(b)<<8+uint16(c), a); cycles += 8; pc += 1 }
func LD_mDE_A() { write(uint16(d)<<8+uint16(e), a); cycles += 8; pc += 1 }
func LDI_mHL_A() {
	hl := uint16(h)<<8 + uint16(l)
	write(hl, a)
	shl := int32(hl) + 1
	h = uint8((shl >> 8) & 0xff)
	l = uint8(shl & 0xff)
	cycles += 8
	pc += 1
}
func LDD_mHL_A() {
	hl := uint16(h)<<8 + uint16(l)
	write(hl, a)
	shl := int32(hl) - 1
	h = uint8((shl >> 8) & 0xff)
	l = uint8(shl & 0xff)
	cycles += 8
	pc += 1
}

func LD_A_mBC() { a = read(uint16(b)<<8 + uint16(c)); cycles += 8; pc += 1 }
func LD_A_mDE() { a = read(uint16(d)<<8 + uint16(e)); cycles += 8; pc += 1 }
func LDI_A_mHL() {
	hl := uint16(h)<<8 + uint16(l)
	a = read(hl)
	shl := int32(hl) + 1
	h = uint8((shl >> 8) & 0xff)
	l = uint8(shl & 0xff)
	cycles += 8
	pc += 1
}
func LDD_A_mHL() {
	hl := uint16(h)<<8 + uint16(l)
	a = read(hl)
	shl := int32(hl) - 1
	h = uint8((shl >> 8) & 0xff)
	l = uint8(shl & 0xff)
	cycles += 8
	pc += 1
}

func LD_B_N()   { b = read(pc + 1); cycles += 8; pc += 2 }
func LD_C_N()   { c = read(pc + 1); cycles += 8; pc += 2 }
func LD_D_N()   { d = read(pc + 1); cycles += 8; pc += 2 }
func LD_E_N()   { e = read(pc + 1); cycles += 8; pc += 2 }
func LD_H_N()   { h = read(pc + 1); cycles += 8; pc += 2 }
func LD_L_N()   { l = read(pc + 1); cycles += 8; pc += 2 }
func LD_mHL_N() { write(uint16(h)<<8+uint16(l), read(pc+1)); cycles += 12; pc += 2 }
func LD_A_N()   { a = read(pc + 1); cycles += 8; pc += 2 }

func LD_B_B()   { b = b; cycles += 4; pc += 1 }
func LD_B_C()   { b = c; cycles += 4; pc += 1 }
func LD_B_D()   { b = d; cycles += 4; pc += 1 }
func LD_B_E()   { b = e; cycles += 4; pc += 1 }
func LD_B_H()   { b = h; cycles += 4; pc += 1 }
func LD_B_L()   { b = l; cycles += 4; pc += 1 }
func LD_B_mHL() { b = read(uint16(h)<<8 + uint16(l)); cycles += 8; pc += 1 }
func LD_B_A()   { b = a; cycles += 4; pc += 1 }
func LD_C_B()   { c = b; cycles += 4; pc += 1 }
func LD_C_C()   { c = c; cycles += 4; pc += 1 }
func LD_C_D()   { c = d; cycles += 4; pc += 1 }
func LD_C_E()   { c = e; cycles += 4; pc += 1 }
func LD_C_H()   { c = h; cycles += 4; pc += 1 }
func LD_C_L()   { c = l; cycles += 4; pc += 1 }
func LD_C_mHL() { c = read(uint16(h)<<8 + uint16(l)); cycles += 8; pc += 1 }
func LD_C_A()   { c = a; cycles += 4; pc += 1 }
func LD_D_B()   { d = b; cycles += 4; pc += 1 }
func LD_D_C()   { d = c; cycles += 4; pc += 1 }
func LD_D_D()   { d = d; cycles += 4; pc += 1 }
func LD_D_E()   { d = e; cycles += 4; pc += 1 }
func LD_D_H()   { d = h; cycles += 4; pc += 1 }
func LD_D_L()   { d = l; cycles += 4; pc += 1 }
func LD_D_mHL() { d = read(uint16(h)<<8 + uint16(l)); cycles += 8; pc += 1 }
func LD_D_A()   { d = a; cycles += 4; pc += 1 }
func LD_E_B()   { e = b; cycles += 4; pc += 1 }
func LD_E_C()   { e = c; cycles += 4; pc += 1 }
func LD_E_D()   { e = d; cycles += 4; pc += 1 }
func LD_E_E()   { e = e; cycles += 4; pc += 1 }
func LD_E_H()   { e = h; cycles += 4; pc += 1 }
func LD_E_L()   { e = l; cycles += 4; pc += 1 }
func LD_E_mHL() { e = read(uint16(h)<<8 + uint16(l)); cycles += 8; pc += 1 }
func LD_E_A()   { e = a; cycles += 4; pc += 1 }
func LD_H_B()   { h = b; cycles += 4; pc += 1 }
func LD_H_C()   { h = c; cycles += 4; pc += 1 }
func LD_H_D()   { h = d; cycles += 4; pc += 1 }
func LD_H_E()   { h = e; cycles += 4; pc += 1 }
func LD_H_H()   { h = h; cycles += 4; pc += 1 }
func LD_H_L()   { h = l; cycles += 4; pc += 1 }
func LD_H_mHL() { h = read(uint16(h)<<8 + uint16(l)); cycles += 8; pc += 1 }
func LD_H_A()   { h = a; cycles += 4; pc += 1 }
func LD_L_B()   { l = b; cycles += 4; pc += 1 }
func LD_L_C()   { l = c; cycles += 4; pc += 1 }
func LD_L_D()   { l = d; cycles += 4; pc += 1 }
func LD_L_E()   { l = e; cycles += 4; pc += 1 }
func LD_L_H()   { l = h; cycles += 4; pc += 1 }
func LD_L_L()   { l = l; cycles += 4; pc += 1 }
func LD_L_mHL() { l = read(uint16(h)<<8 + uint16(l)); cycles += 8; pc += 1 }
func LD_L_A()   { l = a; cycles += 4; pc += 1 }
func LD_mHL_B() { write(uint16(h)<<8+uint16(l), b); cycles += 8; pc += 1 }
func LD_mHL_C() { write(uint16(h)<<8+uint16(l), c); cycles += 8; pc += 1 }
func LD_mHL_D() { write(uint16(h)<<8+uint16(l), d); cycles += 8; pc += 1 }
func LD_mHL_E() { write(uint16(h)<<8+uint16(l), e); cycles += 8; pc += 1 }
func LD_mHL_H() { write(uint16(h)<<8+uint16(l), h); cycles += 8; pc += 1 }
func LD_mHL_L() { write(uint16(h)<<8+uint16(l), l); cycles += 8; pc += 1 }
func LD_mHL_A() { write(uint16(h)<<8+uint16(l), a); cycles += 8; pc += 1 }
func LD_A_B()   { a = b; cycles += 4; pc += 1 }
func LD_A_C()   { a = c; cycles += 4; pc += 1 }
func LD_A_D()   { a = d; cycles += 4; pc += 1 }
func LD_A_E()   { a = e; cycles += 4; pc += 1 }
func LD_A_H()   { a = h; cycles += 4; pc += 1 }
func LD_A_L()   { a = l; cycles += 4; pc += 1 }
func LD_A_mHL() { a = read(uint16(h)<<8 + uint16(l)); cycles += 8; pc += 1 }
func LD_A_A()   { a = a; cycles += 4; pc += 1 }

func LDH_mN_A() { write(0xff00+uint16(read(pc+1)), a); cycles += 12; pc += 2 }
func LDH_A_mN() { a = read(0xff00 + uint16(read(pc+1))); cycles += 12; pc += 2 }
func LDH_mC_A() { write(0xff00+uint16(c), a); cycles += 8; pc += 1 }
func LDH_A_mC() { a = read(0xff00 + uint16(c)); cycles += 8; pc += 1 }

func LD_mNN_A() { write(uint16(read(pc+2))<<8+uint16(read(pc+1)), a); cycles += 16; pc += 3 }
func LD_A_mNN() { a = read(uint16(read(pc+2))<<8 + uint16(read(pc+1))); cycles += 16; pc += 3 }
func LD_mNN_SP() {
	m := uint16(read(pc+2))<<8 + uint16(read(pc+1))
	write(m, uint8(sp&0xff))
	write(m+1, uint8((sp>>8)&0xff))
	cycles += 20
	pc += 3
}

func PUSH_BC() { write(sp-1, b); write(sp-2, c); sp -= 2; cycles += 16; pc += 1 }
func PUSH_DE() { write(sp-1, d); write(sp-2, e); sp -= 2; cycles += 16; pc += 1 }
func PUSH_HL() { write(sp-1, h); write(sp-2, l); sp -= 2; cycles += 16; pc += 1 }
func PUSH_AF() { write(sp-1, a); write(sp-2, readF()); sp -= 2; cycles += 16; pc += 1 }

func POP_BC() { c = read(sp); b = read(sp + 1); sp += 2; cycles += 12; pc += 1 }
func POP_DE() { e = read(sp); d = read(sp + 1); sp += 2; cycles += 12; pc += 1 }
func POP_HL() { l = read(sp); h = read(sp + 1); sp += 2; cycles += 12; pc += 1 }
func POP_AF() { writeF(read(sp)); a = read(sp + 1); sp += 2; cycles += 12; pc += 1 }

func readF() uint8 {
	var f uint8 = 0
	if fz {
		f |= 0x80
	}
	if fn {
		f |= 0x40
	}
	if fh {
		f |= 0x20
	}
	if fc {
		f |= 0x10
	}
	return f
}

func writeF(f uint8) {
	fz = (f & 0x80) > 0
	fn = (f & 0x40) > 0
	fh = (f & 0x20) > 0
	fc = (f & 0x10) > 0
}

func ADD_A_B() {
	r := int16(a) + int16(b)
	t := uint8(r)
	fz = t == 0
	fn = false
	fh = (a^b^t)&0x10 > 0
	fc = r > 0xff
	a = t
	cycles += 4
	pc += 1
}
func ADD_A_C() {
	r := int16(a) + int16(c)
	t := uint8(r)
	fz = t == 0
	fn = false
	fh = (a^c^t)&0x10 > 0
	fc = r > 0xff
	a = t
	cycles += 4
	pc += 1
}
func ADD_A_D() {
	r := int16(a) + int16(d)
	t := uint8(r)
	fz = t == 0
	fn = false
	fh = (a^d^t)&0x10 > 0
	fc = r > 0xff
	a = t
	cycles += 4
	pc += 1
}
func ADD_A_E() {
	r := int16(a) + int16(e)
	t := uint8(r)
	fz = t == 0
	fn = false
	fh = (a^e^t)&0x10 > 0
	fc = r > 0xff
	a = t
	cycles += 4
	pc += 1
}
func ADD_A_H() {
	r := int16(a) + int16(h)
	t := uint8(r)
	fz = t == 0
	fn = false
	fh = (a^h^t)&0x10 > 0
	fc = r > 0xff
	a = t
	cycles += 4
	pc += 1
}
func ADD_A_L() {
	r := int16(a) + int16(l)
	t := uint8(r)
	fz = t == 0
	fn = false
	fh = (a^l^t)&0x10 > 0
	fc = r > 0xff
	a = t
	cycles += 4
	pc += 1
}
func ADD_A_mHL() {
	mhl := read(uint16(h)<<8 + uint16(l))
	r := int16(a) + int16(mhl)
	t := uint8(r)
	fz = t == 0
	fn = false
	fh = (a^mhl^t)&0x10 > 0
	fc = r > 0xff
	a = t
	cycles += 8
	pc += 1
}
func ADD_A_A() {
	r := int16(a) + int16(a)
	t := uint8(r)
	fz = t == 0
	fn = false
	fh = (a^a^t)&0x10 > 0
	fc = r > 0xff
	a = t
	cycles += 4
	pc += 1
}
func ADD_A_N() {
	n := read(pc + 1)
	r := int16(a) + int16(n)
	t := uint8(r)
	fz = t == 0
	fn = false
	fh = (a^n^t)&0x10 > 0
	fc = r > 0xff
	a = t
	cycles += 8
	pc += 2
}

func ADC_A_B() {
	r := int16(a) + int16(b)
	if fc {
		r += 1
	}
	t := uint8(r)
	fz = t == 0
	fn = false
	fh = (a^b^t)&0x10 > 0
	fc = r > 0xff
	a = t
	cycles += 4
	pc += 1
}
func ADC_A_C() {
	r := int16(a) + int16(c)
	if fc {
		r += 1
	}
	t := uint8(r)
	fz = t == 0
	fn = false
	fh = (a^c^t)&0x10 > 0
	fc = r > 0xff
	a = t
	cycles += 4
	pc += 1
}
func ADC_A_D() {
	r := int16(a) + int16(d)
	if fc {
		r += 1
	}
	t := uint8(r)
	fz = t == 0
	fn = false
	fh = (a^d^t)&0x10 > 0
	fc = r > 0xff
	a = t
	cycles += 4
	pc += 1
}
func ADC_A_E() {
	r := int16(a) + int16(e)
	if fc {
		r += 1
	}
	t := uint8(r)
	fz = t == 0
	fn = false
	fh = (a^e^t)&0x10 > 0
	fc = r > 0xff
	a = t
	cycles += 4
	pc += 1
}
func ADC_A_H() {
	r := int16(a) + int16(h)
	if fc {
		r += 1
	}
	t := uint8(r)
	fz = t == 0
	fn = false
	fh = (a^h^t)&0x10 > 0
	fc = r > 0xff
	a = t
	cycles += 4
	pc += 1
}
func ADC_A_L() {
	r := int16(a) + int16(l)
	if fc {
		r += 1
	}
	t := uint8(r)
	fz = t == 0
	fn = false
	fh = (a^l^t)&0x10 > 0
	fc = r > 0xff
	a = t
	cycles += 4
	pc += 1
}
func ADC_A_mHL() {
	mhl := read(uint16(h)<<8 + uint16(l))
	r := int16(a) + int16(mhl)
	if fc {
		r += 1
	}
	t := uint8(r)
	fz = t == 0
	fn = false
	fh = (a^mhl^t)&0x10 > 0
	fc = r > 0xff
	a = t
	cycles += 8
	pc += 1
}
func ADC_A_A() {
	r := int16(a) + int16(a)
	if fc {
		r += 1
	}
	t := uint8(r)
	fz = t == 0
	fn = false
	fh = (a^a^t)&0x10 > 0
	fc = r > 0xff
	a = t
	cycles += 4
	pc += 1
}
func ADC_A_N() {
	n := read(pc + 1)
	r := int16(a) + int16(n)
	if fc {
		r += 1
	}
	t := uint8(r)
	fz = t == 0
	fn = false
	fh = (a^n^t)&0x10 > 0
	fc = r > 0xff
	a = t
	cycles += 8
	pc += 2
}

func SUB_A_B() {
	r := int16(a) - int16(b)
	t := uint8(r)
	fz = t == 0
	fn = true
	fh = (a^b^t)&0x10 > 0
	fc = r < 0
	a = t
	cycles += 4
	pc += 1
}
func SUB_A_C() {
	r := int16(a) - int16(c)
	t := uint8(r)
	fz = t == 0
	fn = true
	fh = (a^c^t)&0x10 > 0
	fc = r < 0
	a = t
	cycles += 4
	pc += 1
}
func SUB_A_D() {
	r := int16(a) - int16(d)
	t := uint8(r)
	fz = t == 0
	fn = true
	fh = (a^d^t)&0x10 > 0
	fc = r < 0
	a = t
	cycles += 4
	pc += 1
}
func SUB_A_E() {
	r := int16(a) - int16(e)
	t := uint8(r)
	fz = t == 0
	fn = true
	fh = (a^e^t)&0x10 > 0
	fc = r < 0
	a = t
	cycles += 4
	pc += 1
}
func SUB_A_H() {
	r := int16(a) - int16(h)
	t := uint8(r)
	fz = t == 0
	fn = true
	fh = (a^h^t)&0x10 > 0
	fc = r < 0
	a = t
	cycles += 4
	pc += 1
}
func SUB_A_L() {
	r := int16(a) - int16(l)
	t := uint8(r)
	fz = t == 0
	fn = true
	fh = (a^l^t)&0x10 > 0
	fc = r < 0
	a = t
	cycles += 4
	pc += 1
}
func SUB_A_mHL() {
	mhl := read(uint16(h)<<8 + uint16(l))
	r := int16(a) - int16(mhl)
	t := uint8(r)
	fz = t == 0
	fn = true
	fh = (a^mhl^t)&0x10 > 0
	fc = r < 0
	a = t
	cycles += 8
	pc += 1
}
func SUB_A_A() {
	fz = true
	fn = true
	fh = false
	fc = false
	a = 0
	cycles += 4
	pc += 1
}
func SUB_A_N() {
	n := read(pc + 1)
	r := int16(a) - int16(n)
	t := uint8(r)
	fz = t == 0
	fn = true
	fh = (a^n^t)&0x10 > 0
	fc = r < 0
	a = t
	cycles += 8
	pc += 2
}

func SBC_A_B() {
	r := int16(a) - int16(b)
	if fc {
		r -= 1
	}
	t := uint8(r)
	fz = t == 0
	fn = true
	fh = (a^b^t)&0x10 > 0
	fc = r < 0
	a = t
	cycles += 4
	pc += 1
}
func SBC_A_C() {
	r := int16(a) - int16(c)
	if fc {
		r -= 1
	}
	t := uint8(r)
	fz = t == 0
	fn = true
	fh = (a^c^t)&0x10 > 0
	fc = r < 0
	a = t
	cycles += 4
	pc += 1
}
func SBC_A_D() {
	r := int16(a) - int16(d)
	if fc {
		r -= 1
	}
	t := uint8(r)
	fz = t == 0
	fn = true
	fh = (a^d^t)&0x10 > 0
	fc = r < 0
	a = t
	cycles += 4
	pc += 1
}
func SBC_A_E() {
	r := int16(a) - int16(e)
	if fc {
		r -= 1
	}
	t := uint8(r)
	fz = t == 0
	fn = true
	fh = (a^e^t)&0x10 > 0
	fc = r < 0
	a = t
	cycles += 4
	pc += 1
}
func SBC_A_H() {
	r := int16(a) - int16(h)
	if fc {
		r -= 1
	}
	t := uint8(r)
	fz = t == 0
	fn = true
	fh = (a^h^t)&0x10 > 0
	fc = r < 0
	a = t
	cycles += 4
	pc += 1
}
func SBC_A_L() {
	r := int16(a) - int16(l)
	if fc {
		r -= 1
	}
	t := uint8(r)
	fz = t == 0
	fn = true
	fh = (a^l^t)&0x10 > 0
	fc = r < 0
	a = t
	cycles += 4
	pc += 1
}
func SBC_A_mHL() {
	mhl := read(uint16(h)<<8 + uint16(l))
	r := int16(a) - int16(mhl)
	if fc {
		r -= 1
	}
	t := uint8(r)
	fz = t == 0
	fn = true
	fh = (a^mhl^t)&0x10 > 0
	fc = r < 0
	a = t
	cycles += 8
	pc += 1
}
func SBC_A_A() {
	fz = false
	fn = true
	fh = true
	fc = true
	if fc {
		a = 0xff
	} else {
		a = 0
	}
	cycles += 4
	pc += 1
}
func SBC_A_N() {
	n := read(pc + 1)
	r := int16(a) - int16(n)
	if fc {
		r -= 1
	}
	t := uint8(r)
	fz = t == 0
	fn = true
	fh = (a^n^t)&0x10 > 0
	fc = r < 0
	a = t
	cycles += 8
	pc += 2
}

func AND_B() {
	a = a & b
	fz = a == 0
	fn = false
	fh = true
	fc = false
	cycles += 4
	pc += 1
}
func AND_C() {
	a = a & c
	fz = a == 0
	fn = false
	fh = true
	fc = false
	cycles += 4
	pc += 1
}
func AND_D() {
	a = a & d
	fz = a == 0
	fn = false
	fh = true
	fc = false
	cycles += 4
	pc += 1
}
func AND_E() {
	a = a & e
	fz = a == 0
	fn = false
	fh = true
	fc = false
	cycles += 4
	pc += 1
}
func AND_H() {
	a = a & h
	fz = a == 0
	fn = false
	fh = true
	fc = false
	cycles += 4
	pc += 1
}
func AND_L() {
	a = a & l
	fz = a == 0
	fn = false
	fh = true
	fc = false
	cycles += 4
	pc += 1
}
func AND_mHL() {
	a = a & read(uint16(h)<<8+uint16(l))
	fz = a == 0
	fn = false
	fh = true
	fc = false
	cycles += 8
	pc += 1
}
func AND_A() {
	fz = a == 0
	fn = false
	fh = true
	fc = false
	cycles += 4
	pc += 1
}
func AND_N() {
	a = a & read(pc+1)
	fz = a == 0
	fn = false
	fh = true
	fc = false
	cycles += 8
	pc += 2
}

func XOR_B() {
	a = a ^ b
	fz = a == 0
	fn = false
	fh = false
	fc = false
	cycles += 4
	pc += 1
}
func XOR_C() {
	a = a ^ c
	fz = a == 0
	fn = false
	fh = false
	fc = false
	cycles += 4
	pc += 1
}
func XOR_D() {
	a = a ^ d
	fz = a == 0
	fn = false
	fh = false
	fc = false
	cycles += 4
	pc += 1
}
func XOR_E() {
	a = a ^ e
	fz = a == 0
	fn = false
	fh = false
	fc = false
	cycles += 4
	pc += 1
}
func XOR_H() {
	a = a ^ h
	fz = a == 0
	fn = false
	fh = false
	fc = false
	cycles += 4
	pc += 1
}
func XOR_L() {
	a = a ^ l
	fz = a == 0
	fn = false
	fh = false
	fc = false
	cycles += 4
	pc += 1
}
func XOR_mHL() {
	a = a ^ read(uint16(h)<<8+uint16(l))
	fz = a == 0
	fn = false
	fh = false
	fc = false
	cycles += 8
	pc += 1
}
func XOR_A() {
	a = 0
	fz = true
	fn = false
	fh = false
	fc = false
	cycles += 4
	pc += 1
}
func XOR_N() {
	a = a ^ read(pc+1)
	fz = a == 0
	fn = false
	fh = false
	fc = false
	cycles += 8
	pc += 2
}

func OR_B() {
	a = a | b
	fz = a == 0
	fn = false
	fh = false
	fc = false
	cycles += 4
	pc += 1
}
func OR_C() {
	a = a | c
	fz = a == 0
	fn = false
	fh = false
	fc = false
	cycles += 4
	pc += 1
}
func OR_D() {
	a = a | d
	fz = a == 0
	fn = false
	fh = false
	fc = false
	cycles += 4
	pc += 1
}
func OR_E() {
	a = a | e
	fz = a == 0
	fn = false
	fh = false
	fc = false
	cycles += 4
	pc += 1
}
func OR_H() {
	a = a | h
	fz = a == 0
	fn = false
	fh = false
	fc = false
	cycles += 4
	pc += 1
}
func OR_L() {
	a = a | l
	fz = a == 0
	fn = false
	fh = false
	fc = false
	cycles += 4
	pc += 1
}
func OR_mHL() {
	a = a | read(uint16(h)<<8+uint16(l))
	fz = a == 0
	fn = false
	fh = false
	fc = false
	cycles += 8
	pc += 1
}
func OR_A() {
	fz = a == 0
	fn = false
	fh = false
	fc = false
	cycles += 4
	pc += 1
}
func OR_N() {
	a = a | read(pc+1)
	fz = a == 0
	fn = false
	fh = false
	fc = false
	cycles += 8
	pc += 2
}

func CP_B() {
	r := int16(a) + ^int16(b) + 1
	t := uint8(r)
	fz = t == 0
	fn = true
	fh = (a^b^t)&0x10 > 0
	fc = r < 0
	cycles += 4
	pc += 1
}
func CP_C() {
	r := int16(a) + ^int16(c) + 1
	t := uint8(r)
	fz = t == 0
	fn = true
	fh = (a^c^t)&0x10 > 0
	fc = r < 0
	cycles += 4
	pc += 1
}
func CP_D() {
	r := int16(a) + ^int16(d) + 1
	t := uint8(r)
	fz = t == 0
	fn = true
	fh = (a^d^t)&0x10 > 0
	fc = r < 0
	cycles += 4
	pc += 1
}
func CP_E() {
	r := int16(a) + ^int16(e) + 1
	t := uint8(r)
	fz = t == 0
	fn = true
	fh = (a^e^t)&0x10 > 0
	fc = r < 0
	cycles += 4
	pc += 1
}
func CP_H() {
	r := int16(a) + ^int16(h) + 1
	t := uint8(r)
	fz = t == 0
	fn = true
	fh = (a^h^t)&0x10 > 0
	fc = r < 0
	cycles += 4
	pc += 1
}
func CP_L() {
	r := int16(a) + ^int16(l) + 1
	t := uint8(r)
	fz = t == 0
	fn = true
	fh = (a^l^t)&0x10 > 0
	fc = r < 0
	cycles += 4
	pc += 1
}
func CP_mHL() {
	mhl := read(uint16(h)<<8 + uint16(l))
	r := int16(a) + ^int16(mhl) + 1
	t := uint8(r)
	fz = t == 0
	fn = true
	fh = (a^mhl^t)&0x10 > 0
	fc = r < 0
	cycles += 8
	pc += 1
}
func CP_A() {
	fz = true
	fn = true
	fh = false
	fc = false
	cycles += 4
	pc += 1
}
func CP_N() {
	n := read(pc + 1)
	r := int16(a) + ^int16(n) + 1
	t := uint8(r)
	fz = t == 0
	fn = true
	fh = (a^n^t)&0x10 > 0
	fc = r < 0
	cycles += 8
	pc += 2
}

func INC_B() {
	b++
	fz = b == 0
	fn = false
	fh = b&0x0f == 0
	cycles += 4
	pc += 1
}
func INC_C() {
	c++
	fz = c == 0
	fn = false
	fh = c&0x0f == 0
	cycles += 4
	pc += 1
}
func INC_D() {
	d++
	fz = d == 0
	fn = false
	fh = d&0x0f == 0
	cycles += 4
	pc += 1
}
func INC_E() {
	e++
	fz = e == 0
	fn = false
	fh = e&0x0f == 0
	cycles += 4
	pc += 1
}
func INC_H() {
	h++
	fz = h == 0
	fn = false
	fh = h&0x0f == 0
	cycles += 4
	pc += 1
}
func INC_L() {
	l++
	fz = l == 0
	fn = false
	fh = l&0x0f == 0
	cycles += 4
	pc += 1
}
func INC_mHL() {
	m := uint16(b)<<8 + uint16(c)
	n := read(m) + 1
	write(m, n)
	fz = n == 0
	fn = false
	fh = n&0x0f == 0
	cycles += 12
	pc += 1
}
func INC_A() {
	a++
	fz = a == 0
	fn = false
	fh = a&0x0f == 0
	cycles += 4
	pc += 1
}

func DEC_B() {
	b--
	fz = b == 0
	fn = true
	fh = b&0x0f == 0x0f
	cycles += 4
	pc += 1
}
func DEC_C() {
	c--
	fz = c == 0
	fn = true
	fh = c&0x0f == 0x0f
	cycles += 4
	pc += 1
}
func DEC_D() {
	d--
	fz = d == 0
	fn = true
	fh = d&0x0f == 0x0f
	cycles += 4
	pc += 1
}
func DEC_E() {
	e--
	fz = e == 0
	fn = true
	fh = e&0x0f == 0x0f
	cycles += 4
	pc += 1
}
func DEC_H() {
	h--
	fz = h == 0
	fn = true
	fh = h&0x0f == 0x0f
	cycles += 4
	pc += 1
}
func DEC_L() {
	l--
	fz = l == 0
	fn = true
	fh = l&0x0f == 0x0f
	cycles += 4
	pc += 1
}
func DEC_mHL() {
	m := uint16(b)<<8 + uint16(c)
	n := read(m) - 1
	write(m, n)
	fz = n == 0
	fn = true
	fh = n&0x0f == 0x0f
	cycles += 12
	pc += 1
}
func DEC_A() {
	a--
	fz = a == 0
	fn = true
	fh = a&0x0f == 0x0f
	cycles += 4
	pc += 1
}

func ADD_HL_BC() {
	bc := int32(b)<<8 + int32(c)
	hl := int32(h)<<8 + int32(l)
	r := bc + hl
	fn = false
	fh = (bc^hl^r)&0x1000 > 0
	fc = r > 0xffff
	h = uint8(r >> 8)
	l = uint8(r)
	cycles += 8
	pc += 1
}
func ADD_HL_DE() {
	de := int32(d)<<8 + int32(e)
	hl := int32(h)<<8 + int32(l)
	r := de + hl
	fn = false
	fh = (de^hl^r)&0x1000 > 0
	fc = r > 0xffff
	h = uint8(r >> 8)
	l = uint8(r)
	cycles += 8
	pc += 1
}
func ADD_HL_HL() {
	hl := int32(h)<<8 + int32(l)
	r := hl + hl
	fn = false
	fh = r&0x1000 > 0
	fc = r > 0xffff
	h = uint8(r >> 8)
	l = uint8(r)
	cycles += 8
	pc += 1
}
func ADD_HL_SP() {
	hl := int32(h)<<8 + int32(l)
	r := int32(sp) + hl
	fn = false
	fh = (int32(sp)^hl^r)&0x1000 > 0
	fc = r > 0xffff
	h = uint8(r >> 8)
	l = uint8(r)
	cycles += 8
	pc += 1
}

func ADD_SP_N() {
	wsp := int32(sp)
	wn := int32(read(pc + 1))
	r := wsp + wn
	fz = false
	fn = false
	fh = (wsp^wn^r)&0x10 > 0
	fc = (wsp^wn^r)&0x100 > 0
	sp = uint16(r)
	cycles += 16
	pc += 2
}
func LD_HL_SP_N() {
	wsp := int32(sp)
	wn := int32(read(pc + 1))
	r := wsp + wn
	fz = false
	fn = false
	fh = (wsp^wn^r)&0x10 > 0
	fc = (wsp^wn^r)&0x100 > 0
	h = uint8(r >> 8)
	l = uint8(r)
	cycles += 12
	pc += 2
}
