package main

import (
	"fmt"
	"github.com/djhworld/gomeboycolor/utils"
	"log"
)

type Register byte
type Word uint16

//flags
const (
	_ = iota
	C
	H
	N
	Z
)

type Registers struct {
	A byte
	B byte
	C byte
	D byte
	E byte
	H byte
	L byte
	F byte // Flags Register
}

//See ZILOG z80 cpu manual p.80  (http://www.zilog.com/docs/z80/um0080.pdf)
type Clock struct {
	m Word
	t Word
}

func (c *Clock) Set(m, t Word) {
	c.m, c.t = m, t
}

func (c *Clock) Reset() {
	c.m, c.t = 0, 0
}

func (c *Clock) String() string {
	return fmt.Sprintf("[M: %X, T: %X]", c.m, c.t)
}

type Z80 struct {
	PC             Word // Program Counter
	SP             Word // Stack Pointer
	R              Registers
	MachineCycles  Clock
	LastInstrCycle Clock
	mmu            MMU
}

func NewCPU(m MMU) *Z80 {
	cpu := new(Z80)
	cpu.mmu = m
	cpu.Reset()

	//TODO: startup. additional setup here

	return cpu
}

func (cpu *Z80) Reset() {
	cpu.PC = 0
	cpu.SP = 0
	cpu.R.A = 0
	cpu.R.B = 0
	cpu.R.C = 0
	cpu.R.D = 0
	cpu.R.E = 0
	cpu.R.F = 0
	cpu.R.H = 0
	cpu.R.L = 0
	cpu.MachineCycles.Reset()
	cpu.LastInstrCycle.Reset()
}

func (cpu *Z80) String() string {
	var flags string = ""

	if cpu.R.F != 0 {
		if cpu.IsFlagSet(Z) {
			flags += "Z "
		}

		if cpu.IsFlagSet(N) {
			flags += "N "
		}

		if cpu.IsFlagSet(H) {
			flags += "H "
		}

		if cpu.IsFlagSet(C) {
			flags += "C "
		}
	} else {
		flags += "None Set"
	}

	return fmt.Sprintf("\nZ80 CPU\n") +
		fmt.Sprintf("--------------------------------------------------------\n") +
		fmt.Sprintf("\tPC		= %X\n", cpu.PC) +
		fmt.Sprintf("\tSP		= %X\n", cpu.SP) +
		fmt.Sprintf("\tLast Cycle	= %v\n", cpu.LastInstrCycle.String()) +
		fmt.Sprintf("\tMachine Cycles	= %v\n", cpu.MachineCycles.String()) +
		fmt.Sprintf("\tFlags		= %v\n", flags) +
		fmt.Sprintf("\n\tRegisters\n") +
		fmt.Sprintf("\tA:%X\tB:%X\tC:%X\tD:%X\n\tE:%X\tF:%X\tH:%X\tL:%X", cpu.R.A, cpu.R.B, cpu.R.C, cpu.R.D, cpu.R.E, cpu.R.F, cpu.R.H, cpu.R.L) +
		fmt.Sprintf("\n--------------------------------------------------------\n\n")
}

func (cpu *Z80) ResetFlag(flag int) {
	switch flag {
	case Z:
		cpu.R.F = cpu.R.F &^ 0x80
	case N:
		cpu.R.F = cpu.R.F &^ 0x40
	case H:
		cpu.R.F = cpu.R.F &^ 0x20
	case C:
		cpu.R.F = cpu.R.F &^ 0x10
	default:
		log.Fatalf("Unknown flag %c", flag)
	}
}

func (cpu *Z80) SetFlag(flag int) {
	switch flag {
	case Z:
		cpu.R.F = cpu.R.F ^ 0x80
	case N:
		cpu.R.F = cpu.R.F ^ 0x40
	case H:
		cpu.R.F = cpu.R.F ^ 0x20
	case C:
		cpu.R.F = cpu.R.F ^ 0x10
	default:
		log.Fatalf("Unknown flag %c", flag)
	}
}

func (cpu *Z80) IsFlagSet(flag int) bool {
	switch flag {
	case Z:
		return cpu.R.F&0x80 == 0x80
	case N:
		return cpu.R.F&0x40 == 0x40
	case H:
		return cpu.R.F&0x20 == 0x20
	case C:
		return cpu.R.F&0x10 == 0x10
	default:
		log.Fatalf("Unknown flag %c", flag)
	}
	return false
}

func (cpu *Z80) IncrementPC(by Word) {
	cpu.PC += by
}

func (cpu *Z80) Dispatch(Opcode byte) {
	switch Opcode {
	case 0x00: //NOP
		cpu.NOP()
	case 0x76: //HALT
		cpu.HALT()
	case 0xF3: //DI
		cpu.DI()
	case 0xFB: //EI
		cpu.EI()

	case 0xA7: //AND A, A
		cpu.AndA_r(&cpu.R.A)
	case 0xA0: //AND A, B
		cpu.AndA_r(&cpu.R.B)
	case 0xA1: //AND A, C
		cpu.AndA_r(&cpu.R.C)
	case 0xA2: //AND A, D
		cpu.AndA_r(&cpu.R.D)
	case 0xA3: //AND A, E
		cpu.AndA_r(&cpu.R.E)
	case 0xA4: //AND A, H
		cpu.AndA_r(&cpu.R.H)
	case 0xA5: //AND A, L
		cpu.AndA_r(&cpu.R.L)
	case 0xA6: //AND A, (HL)
		cpu.AndA_hl()
	case 0xE6: //AND A, n
		cpu.AndA_n()

	case 0xB7: //OR A, A
		cpu.OrA_r(&cpu.R.A)
	case 0xB0: //OR A, B
		cpu.OrA_r(&cpu.R.B)
	case 0xB1: //OR A, C
		cpu.OrA_r(&cpu.R.C)
	case 0xB2: //OR A, D
		cpu.OrA_r(&cpu.R.D)
	case 0xB3: //OR A, E
		cpu.OrA_r(&cpu.R.E)
	case 0xB4: //OR A, H
		cpu.OrA_r(&cpu.R.H)
	case 0xB5: //OR A, L
		cpu.OrA_r(&cpu.R.L)
	case 0xB6: //OR A,(HL)
		cpu.OrA_hl()
	case 0xF6: //OR A, n
		cpu.OrA_n()

	case 0xAF: //XOR A, A
		cpu.XorA_r(&cpu.R.A)
	case 0xA8: //XOR A, B
		cpu.XorA_r(&cpu.R.B)
	case 0xA9: //XOR A, C
		cpu.XorA_r(&cpu.R.C)
	case 0xAA: //XOR A, D
		cpu.XorA_r(&cpu.R.D)
	case 0xAB: //XOR A, E
		cpu.XorA_r(&cpu.R.E)
	case 0xAC: //XOR A, H
		cpu.XorA_r(&cpu.R.H)
	case 0xAD: //XOR A, L
		cpu.XorA_r(&cpu.R.L)
	case 0xAE: //XOR A,(HL)
		cpu.XorA_hl()
	case 0xEE: //XOR A, n
		cpu.XorA_n()

	case 0x97: // SUB A, A
		cpu.SubA_r(&cpu.R.A)
	case 0x90: // SUB A, B
		cpu.SubA_r(&cpu.R.B)
	case 0x91: // SUB A, C
		cpu.SubA_r(&cpu.R.C)
	case 0x92: // SUB A, D
		cpu.SubA_r(&cpu.R.D)
	case 0x93: // SUB A, E
		cpu.SubA_r(&cpu.R.E)
	case 0x94: // SUB A, H
		cpu.SubA_r(&cpu.R.H)
	case 0x95: // SUB A, L
		cpu.SubA_r(&cpu.R.L)
	case 0x96: //SUB A, (HL)
		cpu.SubA_hl()
	case 0xD6: //SUB A, n
		cpu.SubA_n()

	case 0x9F: // SBC A, A
		//TODO: Implement
		log.Fatal("Unimplemnted")
		//cpu.SubAC_r(&cpu.R.A)
	case 0x98: // SBC A, B
		//TODO: Implement
		log.Fatal("Unimplemnted")
		//cpu.SubAC_r(&cpu.R.B)
	case 0x99: // SBC A, C
		//TODO: Implement
		log.Fatal("Unimplemnted")
		//cpu.SubAC_r(&cpu.R.C)
	case 0x9A: // SBC A, D
		//TODO: Implement
		log.Fatal("Unimplemnted")
		//cpu.SubAC_r(&cpu.R.D)
	case 0x9B: // SBC A, E
		//TODO: Implement
		log.Fatal("Unimplemnted")
		//cpu.SubAC_r(&cpu.R.E)
	case 0x9C: // SBC A, H
		//TODO: Implement
		log.Fatal("Unimplemnted")
		//cpu.SubAC_r(&cpu.R.H)
	case 0x9D: // SBC A, L
		//TODO: Implement
		log.Fatal("Unimplemnted")
		//cpu.SubAC_r(&cpu.R.L)
	case 0x9E: //SBC A, (HL)
		//TODO: Implement
		log.Fatal("Unimplemnted")
		//cpu.SubAC_hl()
	case 0xDE: //SBC A, n
		//TODO: Implement
		log.Fatal("Unimplemnted")

	case 0xF5: //PUSH AF
		cpu.Push_nn(&cpu.R.A, &cpu.R.F)
	case 0xC5: //PUSH BC
		cpu.Push_nn(&cpu.R.B, &cpu.R.C)
	case 0xD5: //PUSH DE
		cpu.Push_nn(&cpu.R.D, &cpu.R.E)
	case 0xE5: //PUSH HL
		cpu.Push_nn(&cpu.R.H, &cpu.R.L)

	case 0x8F: //ADC A, A
		cpu.AddCA_r(&cpu.R.A)
	case 0x88: //ADC A, B
		cpu.AddCA_r(&cpu.R.B)
	case 0x89: //ADC A, C
		cpu.AddCA_r(&cpu.R.C)
	case 0x8A: //ADC A, D
		cpu.AddCA_r(&cpu.R.D)
	case 0x8B: //ADC A, E
		cpu.AddCA_r(&cpu.R.E)
	case 0x8C: //ADC A, H
		cpu.AddCA_r(&cpu.R.H)
	case 0x8D: //ADC A, L
		cpu.AddCA_r(&cpu.R.L)
	case 0x8E: //ADC A, (HL)
		cpu.AddCA_hl()
	case 0xCE: //ADC A, n
		cpu.AddCA_n()

	case 0xF1: //POP AF
		cpu.Pop_nn(&cpu.R.A, &cpu.R.F)
	case 0xC1: //POP BC
		cpu.Pop_nn(&cpu.R.B, &cpu.R.C)
	case 0xD1: //POP DE
		cpu.Pop_nn(&cpu.R.D, &cpu.R.E)
	case 0xE1: //POP HL
		cpu.Pop_nn(&cpu.R.H, &cpu.R.L)

	case 0x08: //LD nn, SP
		cpu.LDnn_SP()
	case 0x3E: //LD A, n
		cpu.LDrn(&cpu.R.A)
	case 0x06: //LD B,n
		cpu.LDrn(&cpu.R.B)
	case 0x0E: //LD C,n
		cpu.LDrn(&cpu.R.C)
	case 0x16: //LD D,n
		cpu.LDrn(&cpu.R.D)
	case 0x1E: //LD E,n
		cpu.LDrn(&cpu.R.E)
	case 0x26: //LD H,n
		cpu.LDrn(&cpu.R.H)
	case 0x2E: //LD L,n
		cpu.LDrn(&cpu.R.L)

	case 0x7E: //LD A, (HL)
		cpu.LDr_hl(&cpu.R.A)
	case 0x0A: //LD A, (BC)
		cpu.LDr_bc(&cpu.R.A)
	case 0x1A: //LD A, (DE)
		cpu.LDr_de(&cpu.R.A)
	case 0xFA: //LD A, (nn)
		cpu.LDr_nn(&cpu.R.A)

	case 0x7F: //LD A, A
		cpu.LDrr(&cpu.R.A, &cpu.R.A)
	case 0x78: //LD A, B
		cpu.LDrr(&cpu.R.A, &cpu.R.B)
	case 0x79: //LD A, C
		cpu.LDrr(&cpu.R.A, &cpu.R.C)
	case 0x7A: //LD A, D
		cpu.LDrr(&cpu.R.A, &cpu.R.D)
	case 0x7B: //LD A, E
		cpu.LDrr(&cpu.R.A, &cpu.R.E)
	case 0x7C: //LD A, H
		cpu.LDrr(&cpu.R.A, &cpu.R.H)
	case 0x7D: //LD A, L
		cpu.LDrr(&cpu.R.A, &cpu.R.L)

	case 0x47: //LD B, A
		cpu.LDrr(&cpu.R.B, &cpu.R.A)
	case 0x40: //LD B, B
		cpu.LDrr(&cpu.R.B, &cpu.R.B)
	case 0x41: //LD B, C
		cpu.LDrr(&cpu.R.B, &cpu.R.C)
	case 0x42: //LD B, D
		cpu.LDrr(&cpu.R.B, &cpu.R.D)
	case 0x43: //LD B, E
		cpu.LDrr(&cpu.R.B, &cpu.R.E)
	case 0x44: //LD B, H
		cpu.LDrr(&cpu.R.B, &cpu.R.H)
	case 0x45: //LD B, L
		cpu.LDrr(&cpu.R.B, &cpu.R.L)
	case 0x46: //LD B, (HL)
		cpu.LDr_hl(&cpu.R.B)

	case 0x4F: //LD C, A
		cpu.LDrr(&cpu.R.C, &cpu.R.A)
	case 0x48: //LD C, B
		cpu.LDrr(&cpu.R.C, &cpu.R.B)
	case 0x49: //LD C, C
		cpu.LDrr(&cpu.R.C, &cpu.R.C)
	case 0x4A: //LD C, D
		cpu.LDrr(&cpu.R.C, &cpu.R.D)
	case 0x4B: //LD C, E
		cpu.LDrr(&cpu.R.C, &cpu.R.E)
	case 0x4C: //LD C, H
		cpu.LDrr(&cpu.R.C, &cpu.R.H)
	case 0x4D: //LD C, L
		cpu.LDrr(&cpu.R.C, &cpu.R.L)
	case 0x4E: //LD C, (HL)
		cpu.LDr_hl(&cpu.R.C)

	case 0x57: //LD D, A
		cpu.LDrr(&cpu.R.D, &cpu.R.A)
	case 0x50: //LD D, B
		cpu.LDrr(&cpu.R.D, &cpu.R.B)
	case 0x51: //LD D, C
		cpu.LDrr(&cpu.R.D, &cpu.R.C)
	case 0x52: //LD D, D
		cpu.LDrr(&cpu.R.D, &cpu.R.D)
	case 0x53: //LD D, E
		cpu.LDrr(&cpu.R.D, &cpu.R.E)
	case 0x54: //LD D, H
		cpu.LDrr(&cpu.R.D, &cpu.R.H)
	case 0x55: //LD D, L
		cpu.LDrr(&cpu.R.D, &cpu.R.L)
	case 0x56: //LD D, (HL)
		cpu.LDr_hl(&cpu.R.D)

	case 0x5F: //LD E, A
		cpu.LDrr(&cpu.R.E, &cpu.R.A)
	case 0x58: //LD E, B
		cpu.LDrr(&cpu.R.E, &cpu.R.B)
	case 0x59: //LD E, C
		cpu.LDrr(&cpu.R.E, &cpu.R.C)
	case 0x5A: //LD E, D
		cpu.LDrr(&cpu.R.E, &cpu.R.D)
	case 0x5B: //LD E, E
		cpu.LDrr(&cpu.R.E, &cpu.R.E)
	case 0x5C: //LD E, H
		cpu.LDrr(&cpu.R.E, &cpu.R.H)
	case 0x5D: //LD E, L
		cpu.LDrr(&cpu.R.E, &cpu.R.L)
	case 0x5E: //LD E, (HL)
		cpu.LDr_hl(&cpu.R.E)

	case 0x67: //LD H, A
		cpu.LDrr(&cpu.R.H, &cpu.R.A)
	case 0x60: //LD H, B
		cpu.LDrr(&cpu.R.H, &cpu.R.B)
	case 0x61: //LD H, C
		cpu.LDrr(&cpu.R.H, &cpu.R.C)
	case 0x62: //LD H, D
		cpu.LDrr(&cpu.R.H, &cpu.R.D)
	case 0x63: //LD H, E
		cpu.LDrr(&cpu.R.H, &cpu.R.E)
	case 0x64: //LD H, H
		cpu.LDrr(&cpu.R.H, &cpu.R.H)
	case 0x65: //LD H, L
		cpu.LDrr(&cpu.R.H, &cpu.R.L)
	case 0x66: //LD H, (HL)
		cpu.LDr_hl(&cpu.R.H)

	case 0x6F: //LD L, A
		cpu.LDrr(&cpu.R.L, &cpu.R.A)
	case 0x68: //LD L, B
		cpu.LDrr(&cpu.R.L, &cpu.R.B)
	case 0x69: //LD L, C
		cpu.LDrr(&cpu.R.L, &cpu.R.C)
	case 0x6A: //LD L, D
		cpu.LDrr(&cpu.R.L, &cpu.R.D)
	case 0x6B: //LD L, E
		cpu.LDrr(&cpu.R.L, &cpu.R.E)
	case 0x6C: //LD L, H
		cpu.LDrr(&cpu.R.L, &cpu.R.H)
	case 0x6D: //LD L, L
		cpu.LDrr(&cpu.R.L, &cpu.R.L)
	case 0x6E: //LD L, (HL)
		cpu.LDr_hl(&cpu.R.L)

	case 0x77: //LD (HL), A
		cpu.LDhl_r(&cpu.R.A)
	case 0x70: //LD (HL), B
		cpu.LDhl_r(&cpu.R.B)
	case 0x71: //LD (HL), C
		cpu.LDhl_r(&cpu.R.C)
	case 0x72: //LD (HL), D
		cpu.LDhl_r(&cpu.R.D)
	case 0x73: //LD (HL), E
		cpu.LDhl_r(&cpu.R.E)
	case 0x74: //LD (HL), H
		cpu.LDhl_r(&cpu.R.H)
	case 0x75: //LD (HL), L
		cpu.LDhl_r(&cpu.R.L)

	case 0x02: //LD (BC), A
		cpu.LDbc_r(&cpu.R.A)
	case 0x12: //LD (DE), A
		cpu.LDde_r(&cpu.R.A)
	case 0xEA: //LD (nn), A
		cpu.LDnn_r(&cpu.R.A)

	case 0x36: //LD (HL), n
		cpu.LDhl_n()

	case 0xF2: //LD A,(C)
		cpu.LDr_ffplusc(&cpu.R.A)
	case 0xE2: //LD (C),A
		cpu.LDffplusc_r(&cpu.R.A)

	case 0x3A: //LDD A, (HL)
		cpu.LDDr_hl(&cpu.R.A)
	case 0x32: //LDD (HL), A
		cpu.LDDhl_r(&cpu.R.A)

	case 0x2A: //LDI A, (HL)
		cpu.LDIr_hl(&cpu.R.A)
	case 0x22: //LDI (HL), A
		cpu.LDIhl_r(&cpu.R.A)

	case 0xE0: //LDH n, r
		cpu.LDHn_r(&cpu.R.A)
	case 0xF0: //LDH r, n
		cpu.LDHr_n(&cpu.R.A)

	case 0x01: //LD BC, nn
		cpu.LDn_nn(&cpu.R.B, &cpu.R.C)
	case 0x11: //LD DE, nn
		cpu.LDn_nn(&cpu.R.D, &cpu.R.E)
	case 0x21: //LD HL, nn
		cpu.LDn_nn(&cpu.R.H, &cpu.R.L)
	case 0x31: //LD SP, nn
		cpu.LDSP_nn()
	case 0xF8: //LDHL SP, n
		cpu.LDHLSP_n()
	case 0xF9: //LD SP, HL
		cpu.LDSP_rr(&cpu.R.H, &cpu.R.L)
	case 0x87: //ADD A, A
		cpu.AddA_r(&cpu.R.A)
	case 0x80: //ADD A, B
		cpu.AddA_r(&cpu.R.B)
	case 0x81: //ADD A, C
		cpu.AddA_r(&cpu.R.C)
	case 0x82: //ADD A, D
		cpu.AddA_r(&cpu.R.D)
	case 0x83: //ADD A, E
		cpu.AddA_r(&cpu.R.E)
	case 0x84: //ADD A, H
		cpu.AddA_r(&cpu.R.H)
	case 0x85: //ADD A, L
		cpu.AddA_r(&cpu.R.L)
	case 0x86: //ADD A,(HL)
		cpu.AddA_hl()
	case 0xC6: //ADD A,#
		cpu.AddA_n()
	default:
		log.Fatalf("Invalid/Unknown instruction %X", Opcode)
	}
}

func (cpu *Z80) Step() {
	var Opcode byte = cpu.mmu.ReadByte(cpu.PC)
	cpu.IncrementPC(1)

	cpu.Dispatch(Opcode)

	cpu.MachineCycles.m += cpu.LastInstrCycle.m
	cpu.MachineCycles.t += cpu.LastInstrCycle.t
	cpu.LastInstrCycle.Reset()
}

// INSTRUCTIONS START
//-----------------------------------------------------------------------

//LD r,n
//Load value (n) from memory address in the PC into register (r) and increment PC by 1 
func (cpu *Z80) LDrn(r *byte) {
	log.Println("LD r,n")
	var value byte = cpu.mmu.ReadByte(cpu.PC)
	cpu.IncrementPC(1)

	*r = value

	//set clock values
	cpu.LastInstrCycle.Set(2, 8)
}

//LD r,r
//Load value from register (r2) into register (r1)
func (cpu *Z80) LDrr(r1 *byte, r2 *byte) {
	log.Println("LD r,r")
	*r1 = *r2

	//set clock values
	cpu.LastInstrCycle.Set(1, 4)
}

//LD r,(HL)
//Load value from memory address located in register pair (HL) into register (r)
func (cpu *Z80) LDr_hl(r *byte) {
	log.Println("LD r,(HL)")

	var HL Word = Word(utils.JoinBytes(cpu.R.H, cpu.R.L))
	var value byte = cpu.mmu.ReadByte(HL)

	*r = value

	//set clock values
	cpu.LastInstrCycle.Set(2, 8)
}

//LD (HL),r
//Load value from register (r) into memory address located at register pair (HL)
func (cpu *Z80) LDhl_r(r *byte) {
	log.Println("LD (HL),r")
	var HL Word = Word(utils.JoinBytes(cpu.R.H, cpu.R.L))
	var value byte = *r

	cpu.mmu.WriteByte(HL, value)

	//set clock values
	cpu.LastInstrCycle.Set(2, 8)
}

//LD (BC),r
//Load value from register (r) into memory address located at register pair (BC)
func (cpu *Z80) LDbc_r(r *byte) {
	log.Println("LD (BC),r")

	var BC Word = Word(utils.JoinBytes(cpu.R.B, cpu.R.C))
	var value byte = *r

	cpu.mmu.WriteByte(BC, value)

	//set clock values
	cpu.LastInstrCycle.Set(2, 8)
}

//LD (DE),r
//Load value from register (r) into memory address located at register pair (DE)
func (cpu *Z80) LDde_r(r *byte) {
	log.Println("LD (DE),r")

	var DE Word = Word(utils.JoinBytes(cpu.R.D, cpu.R.E))
	var value byte = *r

	cpu.mmu.WriteByte(DE, value)

	//set clock values
	cpu.LastInstrCycle.Set(2, 8)
}

//LD nn,r
//Load value from register (r) and put it in memory address (nn) taken from the next 2 bytes of memory from the PC. Increment the PC by 2
func (cpu *Z80) LDnn_r(r *byte) {
	log.Println("LD nn,r")
	var resultAddr Word = cpu.mmu.ReadWord(cpu.PC)
	cpu.IncrementPC(2)

	cpu.mmu.WriteByte(resultAddr, *r)

	cpu.LastInstrCycle.Set(4, 16)
}

//LD (HL),n
//Load the value (n) from the memory address in the PC and put it in the memory address designated by register pair (HL)
func (cpu *Z80) LDhl_n() {
	log.Println("LD (HL),n")
	var HL Word = Word(utils.JoinBytes(cpu.R.H, cpu.R.L))
	var value byte = cpu.mmu.ReadByte(cpu.PC)
	cpu.IncrementPC(1)

	cpu.mmu.WriteByte(HL, value)

	//set clock values
	cpu.LastInstrCycle.Set(3, 12)
}

//LD r, (BC)
//Load the value (n) located in memory address stored in register pair (BC) and put it in register (r)
func (cpu *Z80) LDr_bc(r *byte) {
	log.Println("LD r,(BC)")

	var BC Word = Word(utils.JoinBytes(cpu.R.B, cpu.R.C))
	var value byte = cpu.mmu.ReadByte(BC)

	*r = value

	//set clock values
	cpu.LastInstrCycle.Set(2, 8)
}

//LD r, (DE)
//Load the value (n) located in memory address stored in register pair (DE) and put it in register (r)
func (cpu *Z80) LDr_de(r *byte) {
	log.Println("LD r,(DE)")

	var DE Word = Word(utils.JoinBytes(cpu.R.D, cpu.R.E))
	var value byte = cpu.mmu.ReadByte(DE)

	*r = value

	//set clock values
	cpu.LastInstrCycle.Set(2, 8)
}

//LD r, nn
//Load the value in memory address defined from the next two bytes relative to the PC and store it in register (r). Increment the PC by 2
func (cpu *Z80) LDr_nn(r *byte) {
	log.Println("LD r,(nn)")

	//read 2 bytes from PC
	var nn Word = cpu.mmu.ReadWord(cpu.PC)
	cpu.IncrementPC(2)

	var value byte = cpu.mmu.ReadByte(nn)
	*r = value

	//set clock values
	cpu.LastInstrCycle.Set(4, 16)
}

//LD r,(C)
//Load the value from memory addressed 0xFF00 + value in register C. Store it in register (r)
func (cpu *Z80) LDr_ffplusc(r *byte) {
	log.Println("LD r,(C)")
	var valueAddr Word = 0xFF00 + Word(cpu.R.C)
	*r = cpu.mmu.ReadByte(valueAddr)

	//set clock values
	cpu.LastInstrCycle.Set(2, 8)
}

//LD (C),r
//Load the value from register (r) and store it in memory addressed 0xFF00 + value in register C. 
func (cpu *Z80) LDffplusc_r(r *byte) {
	log.Println("LD (C),r")
	var valueAddr Word = 0xFF00 + Word(cpu.R.C)
	cpu.mmu.WriteByte(valueAddr, *r)

	//set clock values
	cpu.LastInstrCycle.Set(2, 8)
}

//LDD r, (HL)
//Load the value from memory addressed in register pair (HL) and store it in register R. Decrement the HL registers
func (cpu *Z80) LDDr_hl(r *byte) {
	log.Println("LDD r, (HL)")
	var HL Word = Word(utils.JoinBytes(cpu.R.H, cpu.R.L))
	*r = cpu.mmu.ReadByte(HL)

	//decrement HL registers
	cpu.R.L -= 1

	//decrement H too if L is 0xFF
	if cpu.R.L == 0xFF {
		cpu.R.H -= 1
	}

	//set clock timings
	cpu.LastInstrCycle.Set(2, 8)

}

//LDD (HL), r
//Load the value in register (r) and store in memory addressed in register pair (HL). Decrement the HL registers
func (cpu *Z80) LDDhl_r(r *byte) {
	log.Println("LDD (HL), r")
	var HL Word = Word(utils.JoinBytes(cpu.R.H, cpu.R.L))
	cpu.mmu.WriteByte(HL, *r)

	//decrement HL registers
	cpu.R.L -= 1

	//decrement H too if L is 0xFF
	if cpu.R.L == 0xFF {
		cpu.R.H -= 1
	}

	cpu.LastInstrCycle.Set(2, 8)
}

//LDI r, (HL)
//Load the value from memory addressed in register pair (HL) and store it in register R. Increment the HL registers
func (cpu *Z80) LDIr_hl(r *byte) {
	log.Println("LDI r, (HL)")
	var HL Word = Word(utils.JoinBytes(cpu.R.H, cpu.R.L))
	*r = cpu.mmu.ReadByte(HL)

	//increment HL registers
	cpu.R.L += 1

	//increment H too if L is 0x00
	if cpu.R.L == 0x00 {
		cpu.R.H += 1
	}

	//set clock timings
	cpu.LastInstrCycle.Set(2, 8)
}

//LDI (HL), r
//Load the value in register (r) and store in memory addressed in register pair (HL). Increment the HL registers
func (cpu *Z80) LDIhl_r(r *byte) {
	log.Println("LDI (HL), r")
	var HL Word = Word(utils.JoinBytes(cpu.R.H, cpu.R.L))
	cpu.mmu.WriteByte(HL, *r)

	//increment HL registers
	cpu.R.L += 1

	//increment H too if L is 0x00
	if cpu.R.L == 0x00 {
		cpu.R.H += 1
	}

	cpu.LastInstrCycle.Set(2, 8)
}

//LDH n, r
//Load value (n) located in memory address in FF00+PC and store it in register (r). Increment PC by 1
func (cpu *Z80) LDHn_r(r *byte) {
	log.Println("LDH n, r")
	var n byte = cpu.mmu.ReadByte(Word(0xFF00) + cpu.PC)
	*r = n
	cpu.IncrementPC(1)

	cpu.LastInstrCycle.Set(3, 12)
}

//LDH r, n
//Load value (n) in register (r) and store it in memory address FF00+PC. Increment PC by 1
func (cpu *Z80) LDHr_n(r *byte) {
	log.Println("LDH r, n")
	cpu.mmu.WriteByte(Word(0xFF00)+cpu.PC, *r)
	cpu.IncrementPC(1)

	cpu.LastInstrCycle.Set(3, 12)
}

//LD n, nn
func (cpu *Z80) LDn_nn(r1, r2 *byte) {
	log.Printf("LD n, nn")
	var v1 byte = cpu.mmu.ReadByte(cpu.PC)
	var v2 byte = cpu.mmu.ReadByte(cpu.PC + 1)
	cpu.IncrementPC(2)

	//TODO: Is this correct? 
	*r1 = v1
	*r2 = v2

	cpu.LastInstrCycle.Set(3, 12)
}

//LD SP, nn
func (cpu *Z80) LDSP_nn() {
	log.Printf("LD SP, nn")
	var value Word = cpu.mmu.ReadWord(cpu.PC)
	cpu.IncrementPC(2)

	cpu.SP = value

	cpu.LastInstrCycle.Set(3, 12)
}

//LD SP, rr
func (cpu *Z80) LDSP_rr(r1, r2 *byte) {
	log.Printf("LD SP, rr")
	var HL Word = Word(utils.JoinBytes(cpu.R.H, cpu.R.L))
	cpu.SP = HL
	cpu.LastInstrCycle.Set(2, 8)
}

//LDHL SP, n 
func (cpu *Z80) LDHLSP_n() {
	log.Println("LDHL SP,n")
	var n Word = Word(cpu.mmu.ReadByte(cpu.PC))
	cpu.IncrementPC(1)

	var HL Word = cpu.SP + n

	cpu.R.H, cpu.R.L = utils.SplitIntoBytes(uint16(HL))

	//TODO: verify flag settings are correct....
	cpu.ResetFlag(Z)
	cpu.ResetFlag(N)

	//set carry flag
	if cpu.SP+n < cpu.SP {
		cpu.SetFlag(C)
	}

	//set half-carry flag
	if (((cpu.SP & 0xf) + (n & 0xf)) & 0x10) == 0x10 {
		cpu.SetFlag(H)
	}

	cpu.LastInstrCycle.Set(3, 12)
}

//LDHL SP, n 
func (cpu *Z80) LDnn_SP() {
	log.Println("LD nn,SP")

	var nn Word = cpu.mmu.ReadWord(cpu.PC)

	cpu.mmu.WriteWord(nn, cpu.SP)

	cpu.IncrementPC(2)
	cpu.LastInstrCycle.Set(5, 20)
}

//PUSH nn 
//Push register pair nn onto the stack and decrement the SP twice
func (cpu *Z80) Push_nn(r1, r2 *byte) {
	log.Println("PUSH nn")
	cpu.SP--
	cpu.mmu.WriteByte(cpu.SP, *r1)
	cpu.SP--
	cpu.mmu.WriteByte(cpu.SP, *r2)
	cpu.LastInstrCycle.Set(3, 12)
}

//POP nn 
//Pop the stack twice onto register pair nn 
func (cpu *Z80) Pop_nn(r1, r2 *byte) {
	log.Println("Pop nn")
	*r2 = cpu.mmu.ReadByte(cpu.SP)
	cpu.SP++
	*r1 = cpu.mmu.ReadByte(cpu.SP)
	cpu.SP++
	cpu.LastInstrCycle.Set(3, 12)
}

//ADD A,r
//Add the value in register (r) to register A
func (cpu *Z80) AddA_r(r *byte) {
	log.Println("ADD A,r")
	var oldA byte = cpu.R.A
	cpu.R.A += *r

	cpu.ResetFlag(N)

	//set carry flag
	if (oldA + *r) < oldA {
		cpu.SetFlag(C)
	}

	//set zero flag
	if cpu.R.A == 0x00 {
		cpu.SetFlag(Z)
	}

	if (((oldA & 0xf) + (*r & 0xf)) & 0x10) == 0x10 {
		cpu.SetFlag(H)
	}

	//set clock values
	cpu.LastInstrCycle.Set(1, 4)
}

//ADD A,(HL)
//Add the value in memory addressed in register pair (HL) to register A
func (cpu *Z80) AddA_hl() {
	log.Println("ADD A,(HL)")
	var HL Word = Word(utils.JoinBytes(cpu.R.H, cpu.R.L))
	var value byte = cpu.mmu.ReadByte(HL)

	var oldA byte = cpu.R.A
	cpu.R.A += value

	cpu.ResetFlag(N)

	//set carry flag
	if (oldA + value) < oldA {
		cpu.SetFlag(C)
	}

	//set zero flag
	if cpu.R.A == 0x00 {
		cpu.SetFlag(Z)
	}

	//set half carry flag
	if (((oldA & 0xf) + (value & 0xf)) & 0x10) == 0x10 {
		cpu.SetFlag(H)
	}

	//set clock values
	cpu.LastInstrCycle.Set(2, 8)
}

//ADD A,n
//Add the value in memory addressed PC to register A. Increment the PC by 1
func (cpu *Z80) AddA_n() {
	log.Println("ADD A,n")
	var value byte = cpu.mmu.ReadByte(cpu.PC)
	cpu.IncrementPC(1)

	var oldA byte = cpu.R.A
	cpu.R.A += value

	cpu.ResetFlag(N)

	//set carry flag
	if (oldA + value) < oldA {
		cpu.SetFlag(C)
	}

	//set zero flag
	if cpu.R.A == 0x00 {
		cpu.SetFlag(Z)
	}

	//set half carry flag
	if (((oldA & 0xf) + (value & 0xf)) & 0x10) == 0x10 {
		cpu.SetFlag(H)
	}

	//set clock values
	cpu.LastInstrCycle.Set(2, 8)
}

//ADDC A,r
func (cpu *Z80) AddCA_r(r *byte) {
	log.Println("ADDC A, r")
	var oldA byte = cpu.R.A
	var carryFlag byte = 0

	cpu.R.A += *r

	if cpu.IsFlagSet(C) {
		carryFlag = 1
		cpu.R.A += carryFlag
	}

	cpu.ResetFlag(N)

	//set carry flag
	if cpu.R.A < oldA {
		cpu.SetFlag(C)
	}

	//set zero flag
	if cpu.R.A == 0x00 {
		cpu.SetFlag(Z)
	}

	//set half carry flag
	if (((oldA & 0xf) + ((*r + carryFlag) & 0xf)) & 0x10) == 0x10 {
		cpu.SetFlag(H)
	}

	cpu.LastInstrCycle.Set(1, 4)
}

//ADDC A,(HL)
func (cpu *Z80) AddCA_hl() {
	log.Println("ADDC A, (HL)")

	var oldA byte = cpu.R.A
	var carryFlag byte = 0
	var HL Word = Word(utils.JoinBytes(cpu.R.H, cpu.R.L))
	var value byte = cpu.mmu.ReadByte(HL)

	cpu.R.A += value

	if cpu.IsFlagSet(C) {
		carryFlag = 1
		cpu.R.A += carryFlag
	}

	cpu.ResetFlag(N)

	//set carry flag
	if cpu.R.A < oldA {
		//TODO: Should carry flag be reset first if it is already set?
		cpu.SetFlag(C)
	}

	//set zero flag
	if cpu.R.A == 0x00 {
		cpu.SetFlag(Z)
	}

	//set half carry flag
	if (((oldA & 0xf) + ((value + carryFlag) & 0xf)) & 0x10) == 0x10 {
		cpu.SetFlag(H)
	}

	cpu.LastInstrCycle.Set(2, 8)
}

//ADDC A,n
func (cpu *Z80) AddCA_n() {
	log.Println("ADDC A, n")

	var oldA byte = cpu.R.A
	var carryFlag byte = 0
	var value byte = cpu.mmu.ReadByte(cpu.PC)
	cpu.IncrementPC(1)

	cpu.R.A += value

	if cpu.IsFlagSet(C) {
		carryFlag = 1
		cpu.R.A += carryFlag
	}

	cpu.ResetFlag(N)

	//set carry flag
	if cpu.R.A < oldA {
		//TODO: Should carry flag be reset first if it is already set?
		cpu.SetFlag(C)
	}

	//set zero flag
	if cpu.R.A == 0x00 {
		cpu.SetFlag(Z)
	}

	//set half carry flag
	if (((oldA & 0xf) + ((value + carryFlag) & 0xf)) & 0x10) == 0x10 {
		cpu.SetFlag(H)
	}

	cpu.LastInstrCycle.Set(2, 8)
}

//SUB A,r
func (cpu *Z80) SubA_r(r *byte) {
	log.Println("SUB A,r")
	var oldA byte = cpu.R.A

	cpu.R.A -= *r

	//set subtract flag
	cpu.SetFlag(N)

	//set zero flag if needed
	if cpu.R.A == 0x00 {
		cpu.SetFlag(Z)
	}

	//Set Carry flag
	if cpu.R.A > oldA {
		cpu.SetFlag(C)
	}

	//Set half carry flag if needed
	//TODO: don't think this calculation for half carry is quite correct....
	if (((oldA & 0x0F) - (*r & 0x0F)) & 0x10) == 0x10 {
		cpu.SetFlag(H)
	}

	cpu.LastInstrCycle.Set(1, 4)
}

//SUB A,hl
func (cpu *Z80) SubA_hl() {
	log.Println("SUB A,(HL)")
	var oldA byte = cpu.R.A
	var HL Word = Word(utils.JoinBytes(cpu.R.H, cpu.R.L))
	var value byte = cpu.mmu.ReadByte(HL)

	cpu.R.A -= value

	//set subtract flag
	cpu.SetFlag(N)

	//set zero flag if needed
	if cpu.R.A == 0x00 {
		cpu.SetFlag(Z)
	}

	//Set Carry flag
	if cpu.R.A > oldA {
		cpu.SetFlag(C)
	}

	//Set half carry flag if needed
	//TODO: don't think this calculation for half carry is quite correct....
	if (((oldA & 0x0F) - (value & 0x0F)) & 0x10) == 0x10 {
		cpu.SetFlag(H)
	}

	cpu.LastInstrCycle.Set(2, 8)
}

//SUB A,n
func (cpu *Z80) SubA_n() {
	log.Println("SUB A,n")
	var oldA byte = cpu.R.A
	var value byte = cpu.mmu.ReadByte(cpu.PC)
	cpu.IncrementPC(1)

	cpu.R.A -= value

	//set subtract flag
	cpu.SetFlag(N)

	//set zero flag if needed
	if cpu.R.A == 0x00 {
		cpu.SetFlag(Z)
	}

	//Set Carry flag
	if cpu.R.A > oldA {
		cpu.SetFlag(C)
	}

	//Set half carry flag if needed
	//TODO: don't think this calculation for half carry is quite correct....
	if (((oldA & 0x0F) - (value & 0x0F)) & 0x10) == 0x10 {
		cpu.SetFlag(H)
	}

	cpu.LastInstrCycle.Set(2, 8)
}

//AND A, r
func (cpu *Z80) AndA_r(r *byte) {
	log.Println("AND A, r")
	cpu.R.A = cpu.R.A & *r

	cpu.SetFlag(H)
	cpu.ResetFlag(N)
	cpu.ResetFlag(C)

	if cpu.R.A == 0x00 {
		cpu.SetFlag(Z)
	}

	cpu.LastInstrCycle.Set(1, 4)
}

//AND A, (HL)
func (cpu *Z80) AndA_hl() {
	log.Println("AND A, (HL)")
	var HL Word = Word(utils.JoinBytes(cpu.R.H, cpu.R.L))
	var value byte = cpu.mmu.ReadByte(HL)

	cpu.R.A = cpu.R.A & value

	cpu.SetFlag(H)
	cpu.ResetFlag(N)
	cpu.ResetFlag(C)

	if cpu.R.A == 0x00 {
		cpu.SetFlag(Z)
	}

	cpu.LastInstrCycle.Set(2, 8)

}

//AND A, n
func (cpu *Z80) AndA_n() {
	log.Println("AND A, n")
	var value byte = cpu.mmu.ReadByte(cpu.PC)
	cpu.IncrementPC(1)

	cpu.R.A = cpu.R.A & value

	cpu.SetFlag(H)
	cpu.ResetFlag(N)
	cpu.ResetFlag(C)

	if cpu.R.A == 0x00 {
		cpu.SetFlag(Z)
	}

	cpu.LastInstrCycle.Set(2, 8)
}

//OR A, r
func (cpu *Z80) OrA_r(r *byte) {
	log.Println("OR A, r")
	cpu.R.A = cpu.R.A | *r

	cpu.ResetFlag(H)
	cpu.ResetFlag(N)
	cpu.ResetFlag(C)

	if cpu.R.A == 0x00 {
		cpu.SetFlag(Z)
	}

	cpu.LastInstrCycle.Set(1, 4)
}

//OR A, (HL)
func (cpu *Z80) OrA_hl() {
	log.Println("OR A, (HL)")
	var HL Word = Word(utils.JoinBytes(cpu.R.H, cpu.R.L))
	var value byte = cpu.mmu.ReadByte(HL)

	cpu.R.A = cpu.R.A | value

	cpu.ResetFlag(H)
	cpu.ResetFlag(N)
	cpu.ResetFlag(C)

	if cpu.R.A == 0x00 {
		cpu.SetFlag(Z)
	}

	cpu.LastInstrCycle.Set(2, 8)

}

//OR A, n
func (cpu *Z80) OrA_n() {
	log.Println("OR A, n")
	var value byte = cpu.mmu.ReadByte(cpu.PC)
	cpu.IncrementPC(1)

	cpu.R.A = cpu.R.A | value

	cpu.ResetFlag(H)
	cpu.ResetFlag(N)
	cpu.ResetFlag(C)

	if cpu.R.A == 0x00 {
		cpu.SetFlag(Z)
	}

	cpu.LastInstrCycle.Set(2, 8)
}

//XOR A, r
func (cpu *Z80) XorA_r(r *byte) {
	log.Println("XOR A, r")
	cpu.R.A = cpu.R.A ^ *r

	cpu.ResetFlag(H)
	cpu.ResetFlag(N)
	cpu.ResetFlag(C)

	if cpu.R.A == 0x00 {
		cpu.SetFlag(Z)
	}

	cpu.LastInstrCycle.Set(1, 4)
}

//XOR A, (HL)
func (cpu *Z80) XorA_hl() {
	log.Println("XOR A, (HL)")
	var HL Word = Word(utils.JoinBytes(cpu.R.H, cpu.R.L))
	var value byte = cpu.mmu.ReadByte(HL)

	cpu.R.A = cpu.R.A ^ value

	cpu.ResetFlag(H)
	cpu.ResetFlag(N)
	cpu.ResetFlag(C)

	if cpu.R.A == 0x00 {
		cpu.SetFlag(Z)
	}

	cpu.LastInstrCycle.Set(2, 8)

}

//XOR A, n
func (cpu *Z80) XorA_n() {
	log.Println("XOR A, n")
	var value byte = cpu.mmu.ReadByte(cpu.PC)
	cpu.IncrementPC(1)

	cpu.R.A = cpu.R.A ^ value

	cpu.ResetFlag(H)
	cpu.ResetFlag(N)
	cpu.ResetFlag(C)

	if cpu.R.A == 0x00 {
		cpu.SetFlag(Z)
	}

	cpu.LastInstrCycle.Set(2, 8)
}

//NOP
//No operation
func (cpu *Z80) NOP() {
	log.Println("NOP")
	//set clock values
	cpu.LastInstrCycle.Set(1, 4)
}

//HALT
//Halt CPU
func (cpu *Z80) HALT() {
	log.Println("HALT")

	//TODO: Implement
	//set clock values
	cpu.LastInstrCycle.Set(1, 4)
}

//DI
//Disable interrupts 
func (cpu *Z80) DI() {
	log.Println("DI")
	//TODO: Implement 

	//set clock values
	cpu.LastInstrCycle.Set(1, 4)
}

//EI
//Enable interrupts 
func (cpu *Z80) EI() {
	log.Println("EI")
	//TODO: Implement 

	//set clock values
	cpu.LastInstrCycle.Set(1, 4)
}

//-----------------------------------------------------------------------
//INSTRUCTIONS END
