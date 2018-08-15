// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package vm

import (
	"fmt"
	"sync/atomic"
        "log"
        "os"
        //"strconv"
        "encoding/hex"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/params"
)


// Config are the configuration options for the Interpreter
type Config struct {
	// Debug enabled debugging Interpreter options
	Debug bool
	// Tracer is the op code logger
	Tracer Tracer
	// NoRecursion disabled Interpreter call, callcode,
	// delegate call and create.
	NoRecursion bool
	// Enable recording of SHA3/keccak preimages
	EnablePreimageRecording bool
	// JumpTable contains the EVM instruction table. This
	// may be left uninitialised and will be set to the default
	// table.
	JumpTable [256]operation
}

// Interpreter is used to run Ethereum based contracts and will utilise the
// passed environment to query external sources for state information.
// The Interpreter will run the byte code VM based on the passed
// configuration.
type Interpreter struct {
	evm      *EVM
	cfg      Config
	gasTable params.GasTable
	intPool  *intPool

	readOnly   bool   // Whether to throw on stateful modifications
	returnData []byte // Last CALL's return data for subsequent reuse
}

var txPersist ContentFetcher

// NewInterpreter returns a new instance of the Interpreter.
func NewInterpreter(evm *EVM, cfg Config) *Interpreter {
	// We use the STOP instruction whether to see
	// the jump table was initialised. If it was not
	// we'll set the default jump table.
	if !cfg.JumpTable[STOP].valid {
		switch {
		case evm.ChainConfig().IsConstantinople(evm.BlockNumber):
			cfg.JumpTable = constantinopleInstructionSet
		case evm.ChainConfig().IsByzantium(evm.BlockNumber):
			cfg.JumpTable = byzantiumInstructionSet
		case evm.ChainConfig().IsHomestead(evm.BlockNumber):
			cfg.JumpTable = homesteadInstructionSet
		default:
			cfg.JumpTable = frontierInstructionSet
		}
	}

	return &Interpreter{
		evm:      evm,
		cfg:      cfg,
		gasTable: evm.ChainConfig().GasTable(evm.BlockNumber),
		intPool:  newIntPool(),
	}
}

func (in *Interpreter) enforceRestrictions(op OpCode, operation operation, stack *Stack) error {
	if in.evm.chainRules.IsByzantium {
		if in.readOnly {
			// If the interpreter is operating in readonly mode, make sure no
			// state-modifying operation is performed. The 3rd stack item
			// for a call operation is the value. Transferring value from one
			// account to the others means the state is modified and should also
			// return with an error.
			if operation.writes || (op == CALL && stack.Back(2).BitLen() > 0) {
				return errWriteProtection
			}
		}
	}
	return nil
}

// Run loops and evaluates the contract's code with the given input data and returns
// the return byte-slice and an error if one occurred.
//
// It's important to note that any errors returned by the interpreter should be
// considered a revert-and-consume-all-gas operation except for
// errExecutionReverted which means revert-and-keep-gas-left.
func (in *Interpreter) Run(contract *Contract, input []byte) (ret []byte, err error) {
        f, ferr := os.OpenFile("/home/bitnami/interpreter.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
        if ferr != nil {
            log.Fatal("Cannot open file", ferr)
        }
        defer f.Close()
        //_, ferr = f.WriteString("\nCall interpreter.go depth: ")
        //_, ferr = f.WriteString(strconv.Itoa(in.evm.depth))
        //_, ferr = f.WriteString(" with input\n")
        //nbyte := bytes.IndexByte(input, 0)
        //_, ferr = f.WriteString(hex.EncodeToString(input))
        //err1 := ioutil.WriteFile("/home/bitnami/interpreter.out", msgb, 0644)
        //if err1 != nil {
        //    log.Fatal("Cannot create file", err1)
        //}
	_, ferr = f.WriteString("\nIn Interpreter.run() with input: ")
	_, ferr = f.WriteString(hex.EncodeToString(input))

	if isRAA(input) {
                if in.evm.txP != nil {
                    txPersist = in.evm.txP
                } else {
                    in.evm.txP = txPersist
                }
                input = doRAA(input, in.evm.txP)
        }

/*
        if len(input) >= 4 {
            sig := hex.EncodeToString(input[0:4])
            _, ferr = f.WriteString(", Function Signature: ")
            _, ferr = f.WriteString(hex.EncodeToString(input[0:4]))
            if sig == "6c58228a" || sig == "07173de5" || sig == "152227ad" || sig == "19608715"  {
                _, ferr = f.WriteString(", RAA Requested!\n")

          // Get the latest tuple from Analyzer
                if in.evm.txP != nil {
                    txPersist = in.evm.txP
                } else {
                    in.evm.txP = txPersist
                }
		_, ferr = f.WriteString(fmt.Sprintf("interpreter.Run -- txPersist  %T %p\n", txPersist, txPersist))
                _, ferr = f.WriteString(fmt.Sprintf("interpreter.Run -- txP  %T %p\n", in.evm.txP, in.evm.txP))
		var tuple [][]byte = Tuple(in.evm.txP)

		_, ferr = f.WriteString("Tuple:\n")

		if tuple[1] == nil {
			_, ferr = f.WriteString(fmt.Sprintf("Tuple is empty, there are no pending transactions\n"))
		} else {
			_, ferr = f.WriteString(fmt.Sprintf("Address: %x\n", tuple[0]))
			_, ferr = f.WriteString(fmt.Sprintf("Mark: %x\n", tuple[1]))
			_, ferr = f.WriteString(fmt.Sprintf("Val: %x\n", tuple[2]))

			for i := 0; i < 3; i++ {
				for k := 0; k < 32; k++ {
					input[(len(input)-96)+(i*32)+k] = tuple[i][k]
				}
			}

			_, ferr = f.WriteString(fmt.Sprintf("Augmented Input: %s\n", hex.EncodeToString(input)))

		}
           //_, ferr = f.WriteString(Tuple(in.evm.txP))

          // Add tuple to input arguments (RAA step)
          // TODO: need to truncate the last 3x128 characters in input and substitute the encoded tuple (address, mark, value)
          //newinput := common.fromHex(")
            } else {
            _, ferr = f.WriteString(", not serialized.\n")
            }
        }*/
	// Increment the call depth which is restricted to 1024
	in.evm.depth++
	defer func() { in.evm.depth-- }()


	// Reset the previous call's return data. It's unimportant to preserve the old buffer
	// as every returning call will return new data anyway.
	in.returnData = nil

	// Don't bother with the execution if there's no code.
	if len(contract.Code) == 0 {
		return nil, nil
	}

	var (
		op    OpCode        // current opcode
		mem   = NewMemory() // bound memory
		stack = newstack()  // local stack
		// For optimisation reason we're using uint64 as the program counter.
		// It's theoretically possible to go above 2^64. The YP defines the PC
		// to be uint256. Practically much less so feasible.
		pc   = uint64(0) // program counter
		cost uint64
		// copies used by tracer
		pcCopy  uint64 // needed for the deferred Tracer
		gasCopy uint64 // for Tracer to log gas remaining before execution
		logged  bool   // deferred Tracer should ignore already logged steps
	)
	contract.Input = input

	if in.cfg.Debug {
		defer func() {
			if err != nil {
				if !logged {
					in.cfg.Tracer.CaptureState(in.evm, pcCopy, op, gasCopy, cost, mem, stack, contract, in.evm.depth, err)
				} else {
					in.cfg.Tracer.CaptureFault(in.evm, pcCopy, op, gasCopy, cost, mem, stack, contract, in.evm.depth, err)
				}
			}
		}()
	}
        // dump the stack to the log to find the state variables
	_, ferr = f.WriteString("\nStack: ######## \n ")
        if len(stack.data) > 0 {
                for i, val := range stack.data {
                        _, ferr = f.WriteString(fmt.Sprintf("%-3d  %v\n", i, val))
                }
        } else {
                _, ferr = f.WriteString("-- empty --")
        }
	_, ferr = f.WriteString("########\n")
	// The Interpreter main run loop (contextual). This loop runs until either an
	// explicit STOP, RETURN or SELFDESTRUCT is executed, an error occurred during
	// the execution of one of the operations or until the done flag is set by the
	// parent context.
	for atomic.LoadInt32(&in.evm.abort) == 0 {
		if in.cfg.Debug {
			// Capture pre-execution values for tracing.
			logged, pcCopy, gasCopy = false, pc, contract.Gas
		}

		// Get the operation from the jump table and validate the stack to ensure there are
		// enough stack items available to perform the operation.
		op = contract.GetOp(pc)
		operation := in.cfg.JumpTable[op]
		if !operation.valid {
			return nil, fmt.Errorf("invalid opcode 0x%x", int(op))
		}
		if err := operation.validateStack(stack); err != nil {
			return nil, err
		}
		// If the operation is valid, enforce and write restrictions
		if err := in.enforceRestrictions(op, operation, stack); err != nil {
			return nil, err
		}

		var memorySize uint64
		// calculate the new memory size and expand the memory to fit
		// the operation
		if operation.memorySize != nil {
			memSize, overflow := bigUint64(operation.memorySize(stack))
			if overflow {
				return nil, errGasUintOverflow
			}
			// memory is expanded in words of 32 bytes. Gas
			// is also calculated in words.
			if memorySize, overflow = math.SafeMul(toWordSize(memSize), 32); overflow {
				return nil, errGasUintOverflow
			}
		}
		// consume the gas and return an error if not enough gas is available.
		// cost is explicitly set so that the capture state defer method can get the proper cost
		cost, err = operation.gasCost(in.gasTable, in.evm, contract, stack, mem, memorySize)
		if err != nil || !contract.UseGas(cost) {
			return nil, ErrOutOfGas
		}
		if memorySize > 0 {
			mem.Resize(memorySize)
		}

		if in.cfg.Debug {
			in.cfg.Tracer.CaptureState(in.evm, pc, op, gasCopy, cost, mem, stack, contract, in.evm.depth, err)
			logged = true
		}
		// execute the operation
		res, err := operation.execute(&pc, in.evm, contract, mem, stack)
		// verifyPool is a build flag. Pool verification makes sure the integrity
		// of the integer pool by comparing values to a default value.
		if verifyPool {
			verifyIntegerPool(in.intPool)
		}
		// if the operation clears the return data (e.g. it has returning data)
		// set the last return to the result of the operation.
		if operation.returns {
			in.returnData = res
		}

		switch {
		case err != nil:
			return nil, err
		case operation.reverts:
			return res, errExecutionReverted
		case operation.halts:
			return res, nil
		case !operation.jumps:
			pc++
		}
	}
	return nil, nil
}
