// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:generate make -C $PRU_SSP/examples/am335x/PRU_RPMsg_Echo_Interrupt0
//go:generate make -C $PRU_SSP/examples/am335x/PRU_RPMsg_Echo_Interrupt1
//go:generate sudo cp $PRU_SSP/examples/am335x/PRU_RPMsg_Echo_Interrupt0/gen/PRU_RPMsg_Echo_Interrupt0.out /lib/firmware/am335x-pru0-echo0-fw
//go:generate sudo cp $PRU_SSP/examples/am335x/PRU_RPMsg_Echo_Interrupt1/gen/PRU_RPMsg_Echo_Interrupt1.out /lib/firmware/am335x-pru1-echo1-fw

package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/aamcrae/pru-rp"
)

var counter sync.WaitGroup

func main() {
	counter.Add(2)
	go run(0, "am335x-pru0-echo0-fw")
	go run(1, "am335x-pru1-echo1-fw")
	counter.Wait()
}

func run(unit int, fw string) {
	var msgs sync.WaitGroup
	p, err := pru.Open(unit)
	if err != nil {
		log.Fatalf("%s", err)
	}
	defer p.Close()
	err = p.Load(fw)
	if err != nil {
		log.Fatalf("Load %s: %v", fw, err)
	}

	p.Callback(func(msg []byte) {
		log.Printf("PRU%d: Rx OK [%s]", unit, msg)
		msgs.Done()
	})
	p.Start(true)
	log.Printf("PRU %d state: %s", unit, p.Status().String())
	for i := 0; i < 10; i++ {
		msgs.Add(1)
		err := p.Send([]byte(fmt.Sprintf("msg %d to PRU%d", i, unit)))
		if err != nil {
			log.Printf("PRU%d: Send error: %v", unit, err)
			msgs.Done()
		}
		time.Sleep(100 * time.Millisecond)
	}
	msgs.Wait()
	counter.Done()
}
