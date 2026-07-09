package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"

	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"

	"encoding/json"
)

type disk_part_collection struct {
	device       string `json:"Device,omitempty"`
	mount_point  string `json:"Mount_Point,omitempty"`
	fstype       string `json:"Fstype,omitempty"`
	total_gb     string `json:"Total_Gb,omitempty"`
	used_gb      string `json:"Used_Gb,omitempty"`
	free_gb      string `json:"Free_Gb,omitempty"`
	percent_used string `json:"Percent_Used,omitempty"`
}

type disk_io_counters struct {
	read_mb     string `json:"Read_Mb,omitempty"`
	write_mb    string `json:"Write_Mb,omitempty"`
	read_count  string `json:"Read_Count,omitempty"`
	write_count string `json:"Write_Count,omitempty"`
}

type disk_collection struct {
	partition    []disk_part_collection      `json:"Partition,omitempty"`
	io_counters  map[string]disk_io_counters `json:"Io_Counters,omitempty"`
	recent_files []recent_file               `json:"Recent_Files,omitempty"`
}

type recent_file struct {
	path     string `json:"Path,omitempty"`
	modified string `json:"Modified,omitempty"`
}

type cpu_collection struct {
	total_use    string            `json:"Total_Use,omitempty"`
	per_core_use string            `json:"Per_Core_Use,omitempty"`
	core_count   string            `json:"Core_Count,omitempty"`
	load_avg     map[string]string `json:"Load_Avg,omitempty"`
	freq         string            `json:"Freq,omitempty"`
}

type mem_proc_obj struct {
	pid         string `json:"Pid,omitempty"`
	name        string `json:"Name,omitempty"`
	mem_percent string `json:"Mem_Percent,omitempty"`
}

type mem_obj struct {
	total_mb     string `json:"Total_Mb,omitempty"`
	used_mb      string `json:"Used_Mb,omitempty"`
	free_mb      string `json:"Free_Mb,omitempty"`
	percent_used string `json:"Percent_Used,omitempty"`
}

type mem_collection struct {
	ram   mem_obj        `json:"Ram"`
	swap  mem_obj        `json:"Swap"`
	procs []mem_proc_obj `json:"Procs,omitempty"`
}

type collect_net struct {
	fd             string `json:"Fd,omitempty"`
	remote_ip      string `json:"Remote_Ip,omitempty"`
	remote_port    string `json:"Remote_Port,omitempty"`
	local_address  string `json:"Local_Address,omitempty"`
	remote_address string `json:"Remote_Address,omitempty"`
	pid            string `json:"Pid,omitempty"`
	status         string `json:"Status,omitempty"`
}

type network_collection struct {
	active_connections     string                       `json:"Active_Connections,omitempty"`
	suspicious_connections []collect_net                `json:"Suspicious_Connections,omitempty"`
	connections            []collect_net                `json:"Connections,omitempty"`
	io_counter             map[string]map[string]string `json:"Io_Counter,omitempty"`
}

type snapshot struct {
	cpu_snap  cpu_collection     `json:"Cpu_Snap"`
	disk_snap disk_collection    `json:"Disk_Snap"`
	mem_snap  mem_collection     `json:"Mem_Snap"`
	net_snap  network_collection `json:"Net_Snap"`
	timestamp string             `json:"timestamp"`
	trigger   string             `json:"trigger"`
}

// collector
func collect_cpu() cpu_collection {
	var cpu_stats cpu_collection

	var usage_tot, _ = cpu.Percent(3*time.Second, false)

	cpu_stats.total_use = fmt.Sprintf("%f", usage_tot[0])

	var usage_per_core, _ = cpu.Percent(3*time.Second, true)

	cpu_stats.per_core_use = fmt.Sprintf("%f", usage_per_core[0])

	var stats, _ = cpu.Info()

	cpu_stats.freq = fmt.Sprintf("%d", stats[0].Mhz)

	cpu_stats.core_count = fmt.Sprintf("%d", stats[0].Cores)

	var load_avgs, _ = load.Avg()

	var load_avg_1 = fmt.Sprintf("%f", load_avgs.Load1)
	var load_avg_5 = fmt.Sprintf("%f", load_avgs.Load5)
	var load_avg_15 = fmt.Sprintf("%f", load_avgs.Load15)

	cpu_stats.load_avg = map[string]string{
		"load1":  load_avg_1,
		"load5":  load_avg_5,
		"load15": load_avg_15,
	}

	return cpu_stats
}

func collect_disk() disk_collection {

	var partitions, _ = disk.Partitions(true)
	var disk_parts []disk_part_collection = make([]disk_part_collection, 0)
	var io_res map[string]disk_io_counters = make(map[string]disk_io_counters)

	for _, p := range partitions {
		var usage, _ = disk.Usage(p.Mountpoint)
		var dp disk_part_collection
		dp.device = p.Device
		dp.fstype = p.Fstype
		dp.free_gb = fmt.Sprintf("%f", usage.Free/1024/1024/1024)
		dp.total_gb = fmt.Sprintf("%f", usage.Total/1024/1024/1024)
		dp.used_gb = fmt.Sprintf("%f", usage.Used/1024/1024/1024)
		dp.percent_used = fmt.Sprintf("%f", usage.UsedPercent)
		dp.mount_point = p.Mountpoint

		disk_parts = append(disk_parts, dp)
	}

	var io_counters, _ = disk.IOCounters()

	var recent_files []recent_file
	var tmp_path, _ = filepath.EvalSymlinks("/tmp")

	var _ = filepath.WalkDir(tmp_path, func(path string, d fs.DirEntry, err error) error {
		if !strings.HasPrefix(path, tmp_path) {
			log.Println("[-] symlink escape blocked!")
			return err
		} else {
			if err != nil {
				return err
			}
			info, err := d.Info()
			if err != nil {
				return err
			}

			var mtime = info.ModTime()
			var r_file recent_file
			r_file.path = path
			r_file.modified = mtime.String()
			recent_files = append(recent_files)
		}

		return nil
	})

	if len(io_counters) > 0 {

		for disk_name, stats := range io_counters {
			var io_res_part disk_io_counters

			var read_mb = stats.ReadBytes / 1024 / 1024
			var write_mb = stats.WriteBytes / 1024 / 1024
			var read_count = stats.ReadCount
			var write_count = stats.WriteCount

			io_res_part.read_mb = fmt.Sprintf("%d", read_mb)
			io_res_part.write_mb = fmt.Sprintf("%d", write_mb)
			io_res_part.read_count = fmt.Sprintf("%d", read_count)
			io_res_part.write_count = fmt.Sprintf("%d", write_count)

			io_res[disk_name] = io_res_part
		}
	}

	var result disk_collection
	result.io_counters = io_res
	result.partition = disk_parts
	result.recent_files = recent_files

	return result
}

func collect_mem() mem_collection {
	log.Println("[+] Starting Memory Collector")
	usage, _ := mem.VirtualMemory()
	swap_usage, _ := mem.SwapMemory()

	//build proc collection
	var proc_use, _ = process.Processes()
	var key_mem_obj map[int]mem_proc_obj
	//var key_mem_obj map[int]float32
	var keys []int

	for k, v := range proc_use {
		var obj mem_proc_obj
		var proc_name, em = v.Name()
		if em == nil {
			obj.name = proc_name
		}
		var pid = v.Pid
		obj.pid = fmt.Sprintf("%d", pid)

		var mem_percent, ep = v.MemoryPercent()
		if ep == nil {
			obj.mem_percent = fmt.Sprintf("%d", mem_percent)
		}

		key_mem_obj[k] = obj
		keys = append(keys, k)
	}

	//sort and then top 5
	sort.Slice(keys, func(i, j int) bool { return key_mem_obj[keys[i]].mem_percent > key_mem_obj[keys[j]].mem_percent })

	var final_mem_ojbs []mem_proc_obj
	//var final_mem_ojbs map[int]float32

	if len(key_mem_obj) >= 5 {
		for i := range 5 {
			var mem_o mem_proc_obj
			mem_o.pid = key_mem_obj[i].pid
			mem_o.name = key_mem_obj[i].name
			mem_o.mem_percent = key_mem_obj[i].mem_percent
			final_mem_ojbs = append(final_mem_ojbs, mem_o)
		}
	} else {
		for i := range len(key_mem_obj) {
			var mem_o mem_proc_obj
			mem_o.pid = key_mem_obj[i].pid
			mem_o.name = key_mem_obj[i].name
			mem_o.mem_percent = key_mem_obj[i].mem_percent
			final_mem_ojbs = append(final_mem_ojbs, mem_o)
		}
	}

	//build ram struct
	var ram_obj mem_obj
	var swap_ram_obj mem_obj

	var total_mb = usage.Total
	var used_mb = usage.Used
	var free_mb = usage.Free

	ram_obj.total_mb = strconv.FormatUint(total_mb, 10)
	ram_obj.free_mb = strconv.FormatUint(free_mb, 10)
	ram_obj.used_mb = strconv.FormatUint(used_mb, 10)

	total_mb = swap_usage.Total
	used_mb = swap_usage.Used
	free_mb = swap_usage.Free

	swap_ram_obj.total_mb = strconv.FormatUint(total_mb, 10)
	swap_ram_obj.used_mb = strconv.FormatUint(used_mb, 10)
	swap_ram_obj.free_mb = strconv.FormatUint(free_mb, 10)

	var result mem_collection

	result.ram = ram_obj
	result.swap = swap_ram_obj
	result.procs = final_mem_ojbs

	return result
}

func collect_network() network_collection {
	var conns, _ = net.Connections("inet")
	var connections []collect_net
	var connections_sus []collect_net
	for _, v := range conns {
		var entry collect_net
		entry.fd = strconv.Itoa(int(v.Fd))
		entry.local_address = fmt.Sprintf("%d:%d", v.Laddr.IP, v.Laddr.Port)
		entry.remote_address = fmt.Sprintf("%d:%d", v.Raddr.IP, v.Raddr.Port)
		entry.status = v.Status
		entry.pid = string(v.Pid)

		connections = append(connections, entry)

		if slices.Contains(SUS_PORTS, entry.remote_port) {
			entry.remote_ip = v.Raddr.IP
			entry.remote_port = strconv.Itoa(int(v.Raddr.Port))

			connections_sus = append(connections_sus, entry)

		}

	}

	//finish io stuff
	var io_counters, _ = net.IOCounters(false)

	//finish building conn collection
	var net_collection network_collection
	net_collection.io_counter = make(map[string]map[string]string)

	for _, v := range io_counters {
		net_collection.io_counter[v.Name]["bytes_sent"] = strconv.FormatUint(v.BytesSent, 10)
		net_collection.io_counter[v.Name]["bytes_recv"] = strconv.FormatUint(v.BytesRecv, 10)
		net_collection.io_counter[v.Name]["packet_sent"] = strconv.FormatUint(v.PacketsSent, 10)
		net_collection.io_counter[v.Name]["packet_recv"] = strconv.FormatUint(v.PacketsRecv, 10)
	}

	net_collection.connections = connections
	net_collection.active_connections = strconv.Itoa(len(connections))
	net_collection.suspicious_connections = connections_sus

	//return con collection
	return net_collection

}

// detectors
func get_cpu() (string, bool) {
	log.Println("[+] Starting CPU Collector")
	var usage, e = cpu.Percent(3*time.Second, false)
	if e != nil {
		log.Println(e)
	}

	var res = false
	if usage[0] > cpu_max {
		res = true
	}

	if Verbose {
		log.Println("[+] Finished CPU Collector")
		log.Printf("CPU: %f", usage[0])
	}
	return fmt.Sprintf("cpu usage %f", usage[0]), res
}

func get_mem() (string, bool) {
	log.Println("[+] Starting Memory Collector")
	usage, _ := mem.VirtualMemory()

	var res = false
	if usage.UsedPercent > mem_max {
		res = true
	}
	if Verbose {
		log.Println("[+] Finished Memory Collector")
		log.Printf("MEM: %f", usage.UsedPercent)

	}
	return fmt.Sprintf("memory usage %f", usage.UsedPercent), res
}

func get_disk() (string, bool) {
	log.Println("[+] Starting Disk Collector")
	var usage, _ = disk.Usage("/")
	var res = false
	if usage.UsedPercent > cpu_max {
		res = true
	}

	if Verbose {
		log.Println("[+] Finished Disk Collector")

	}
	return fmt.Sprintf("disk usage %f\n", usage.UsedPercent), res
}

func get_network() (string, bool) {
	log.Println("[+] Starting Network Collector")
	var cons, _ = net.Connections("all")
	for _, v := range cons {

		if slices.Contains(W_IPS, v.Raddr.String()) {
			if Verbose {
				log.Println("[+] Finished Network Collector")
			}
			return fmt.Sprintf("suspicious address: %s\n", v.Raddr.String()), true
		}

	}

	if Verbose {
		log.Println("[+] Finished Network Collector")
	}

	return "", false
}

func get_procs() (string, bool) {
	log.Println("[+] Starting Processor Collector")

	processes, _ := process.Processes()
	for _, p := range processes {
		name, _ := p.Name()
		if slices.Contains(W_PROCS, name) {
			continue
		}
		mem, _ := p.MemoryPercent()

		if mem > proc_mem_max {

			if Verbose {
				log.Println("[+] Finished Process Collector")
			}

			return fmt.Sprintf("proc: %s high memory: %f\n", name, mem), true
		}
	}

	if Verbose {
		log.Println("[+] Finished Network Collector")
	}
	return "", false

}

func CheckAll() string {
	var cpu_res, c_valid = get_cpu()

	var mem_res, m_valid = get_mem()
	var network_res, n_valid = get_network()
	var process_res, p_valid = get_procs()
	var disk_res, d_valid = get_disk()

	if c_valid {
		return cpu_res
	} else if m_valid {
		return mem_res
	} else if n_valid {
		return network_res
	} else if p_valid {
		return process_res
	} else if d_valid {
		return disk_res
	} else {
		return ""
	}
}

//notifier
// TODO smtp notification?

// reporter
//func output_text(entries [][]string, output string, filename string) {
//
//	log.Println("[+] Saving to file/s text + json")
//	var json_filename = fmt.Sprintf("%s.json")
//	var text_filename = fmt.Sprintf("%s.text")
//
//	//save_to_json(json_filename)
//	//save_to_text(text_filename)
//	//save_snapshot()
//
//	log.Println("[+] Snapshot saved to file/s")
//}

// snapshot
func TakeSnapshot(reason string) snapshot {

	var timestamp = time.Now()

	log.Printf("[!] Snapshot reason: %s", reason)
	log.Printf("[*] Taking snapshot at : %s", timestamp)

	var snap snapshot

	snap.cpu_snap = collect_cpu()
	snap.disk_snap = collect_disk()
	snap.net_snap = collect_network()
	snap.mem_snap = collect_mem()

	snap.timestamp = timestamp.String()

	log.Println("[+] Snapshot captured")

	return snap
}

func save_snapshot(snap snapshot) {

	var ts = time.Now().String()

	save_to_json(snap, ts)

	save_to_text(snap, ts)

}

func save_to_json(snap snapshot, ts string) {
	log.Printf("[+] Saving to json file %s\n", SAVE_DIR)

	var data, e = json.Marshal(snap)

	if e != nil {
		log.Fatalln("Unable to marsha data for output")
	}
	var err = os.WriteFile(fmt.Sprintf("snapshot_%s.json", ts), data, os.ModePerm)

	if err != nil {
		log.Println("Unable to save snapshot to json file!")
	}

	log.Println("[+] Snapshot saved to json file!")

}

func save_to_text(snap snapshot, ts string) {
	log.Printf("[+] Saving to text file %s\n", SAVE_DIR)

	var final_str = ""

	var file, e = os.OpenFile(fmt.Sprintf("snapshot_%s.json", ts), os.O_CREATE, os.ModePerm)

	if e != nil {
		log.Fatalln("unable to creat output text file")
	}

	//start writing data to text file
	final_str += "============================================================\n"
	final_str += "         GO-SECSNAP FORENSIC SNAPSHOT\\n"
	final_str += fmt.Sprintf("         %s\n", snap.timestamp)
	final_str += fmt.Sprintf("         TRIGGER:%s\\n", snap.trigger)
	final_str += "============================================================\n"

	//write cpu info
	final_str += "[CPU]\\n"
	final_str += "--------------------------------------------------\n"
	final_str += fmt.Sprintf("  Total Usage    :%s \\n", snap.cpu_snap.total_use)
	final_str += fmt.Sprintf("  Core Count     :%s \\n", snap.cpu_snap.core_count)
	final_str += fmt.Sprintf("  Per Core       :%s \\n", snap.cpu_snap.per_core_use)
	final_str += fmt.Sprintf("  Frequency MHz  :%s \\n", snap.cpu_snap.freq)
	final_str += fmt.Sprintf("  Load Avg 1m    :%s \\n", snap.cpu_snap.load_avg["load1"])

	//write memory info
	final_str += "[MEMORY]\\n"
	final_str += "--------------------------------------------------\n"
	final_str += fmt.Sprintf("  Total      :%s mb\\n", snap.mem_snap.ram.total_mb)
	final_str += fmt.Sprintf("  Used       :%s mb\\n", snap.mem_snap.ram.percent_used)
	final_str += fmt.Sprintf("  Free       :%s mb\\n", snap.cpu_snap.per_core_use)
	final_str += "  Top Processes  :\\n"
	for _, v := range snap.mem_snap.procs {
		final_str += fmt.Sprintf("    PID %s %s — %s\\n\\n", v.pid, v.name, v.mem_percent)
	}
	//write network info
	final_str += "[NETWORK]\\n"
	final_str += "--------------------------------------------------\n"
	final_str += fmt.Sprintf("  Active Connections      :%s \\n", snap.net_snap.active_connections)
	final_str += fmt.Sprintf("  Suspicious Connections      :%s \\n", len(snap.net_snap.suspicious_connections))
	for _, v := range snap.net_snap.suspicious_connections {
		final_str += fmt.Sprintf("    !! %s:{%s PID %s\n", v.remote_ip, v.remote_port, v.pid)
	}
	for k, v := range snap.net_snap.io_counter {
		final_str += fmt.Sprintf("  IOCounter      :%s \\n", k)
		final_str += fmt.Sprintf("  Bytes Sent      :%s \\n", v["bytes_sent"])
		final_str += fmt.Sprintf("  Bytes Received      :%s \\n\\n", v["bytes_received"])
	}

	//write disk info
	final_str += "[DISK]\\n"
	final_str += "--------------------------------------------------\n"
	for _, v := range snap.disk_snap.partition {
		final_str += fmt.Sprintf("	%s -> %s %s\\n", v.device, v.mount_point, v.fstype)
		final_str += fmt.Sprintf("  Used: %s GB / %s GB :%s \\n", v.used_gb, v.total_gb, v.percent_used)
	}

	final_str += "        DISK IO COUNTERS\\n"
	for _, v := range snap.disk_snap.io_counters {

		final_str += fmt.Sprintf("	READ:  %s MB\\n", v.read_mb)
		final_str += fmt.Sprintf("   WRITE: %s MB\\n", v.write_mb)
	}

	//end writing data to text file
	final_str += "============================================================\n"
	final_str += "                   END SNAPSHOT\\n"
	final_str += "============================================================\n"

	var _, write_error = file.WriteString(final_str)
	if write_error != nil {
		log.Fatalln(write_error)
	}
	log.Println("[+] Snapshot saved to text file!")
}
