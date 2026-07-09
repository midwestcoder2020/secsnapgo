package main

import (
	"log"
	"time"
)

func run_checks() string {
	var should_trigger = ""

	//do checks
	should_trigger = CheckAll()

	return should_trigger
}

func main() {

	var tick = 0

	//
	log.Println("[*] Starting SecSnapGo")
	log.Printf("[+] Monitoring set to: %d seconds", TICK_TIME)
	log.Printf("[+] Whitelisted IPs: %s", W_IPS)
	log.Printf("[+] Whitelisted Processes: %s", W_PROCS)
	log.Printf("[+] Press Ctrl+C to stop")

	//main loop
	for {
		var should_trigger = run_checks()

		if Verbose {
			log.Printf("Trigger Found: %s", should_trigger)
		}

		if (len(should_trigger) > 0) && (tick*MAX_TICK >= COOLDOWN) {
			log.Println("[!] Trigger Found!")
			log.Printf("[!] Trigger %s", should_trigger)
			tick = 0

			//take snapshot
			var snap = TakeSnapshot(should_trigger)

			//save snapshot
			save_snapshot(snap)

		} else {
			log.Println("[!] No Triggers")
			log.Println("[*] Sleeping")
			time.Sleep(MAX_TICK)
			tick += 1
		}

	}

}
