package main

var W_IPS = []string{"127.0.0.0", "0.0.0.0", "192.168.0.1", "192.168.1.1", "10.0.0.1"}
var SUS_PORTS = []string{"4444", "1337", "31337", "9001", "6667"}

var W_PROCS = []string{}

const TICK_TIME = 1
const MAX_TICK = 1
const COOLDOWN = 1

const Verbose = true

const cpu_max = 1
const mem_max = 50
const proc_mem_max = 50
const SAVE_DIR = "./snapshots"
