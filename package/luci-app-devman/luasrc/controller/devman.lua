module("luci.controller.devman", package.seeall)

function index()
	entry({"admin", "network", "devman"}, firstchild(), _("Device Manager"), 50)
	entry({"admin", "network", "devman", "overview"}, template("devman/overview"), _("Overview"), 1)
	entry({"admin", "network", "devman", "rules"}, template("devman/rules"), _("Rules"), 2)
	entry({"admin", "network", "devman", "api_devices"}, call("api_devices"))
	entry({"admin", "network", "devman", "api_block"}, call("api_block"))
	entry({"admin", "network", "devman", "api_limit"}, call("api_limit"))
end

function api_devices()
	local http = require "luci.http"
	local io = require "io"
	local f = io.popen("curl -s http://127.0.0.1:9999/api/devices 2>/dev/null")
	local data = f:read("*a")
	f:close()
	http.prepare_content("application/json")
	http.write(data or "[]")
end

function api_block()
	local http = require "luci.http"
	local dev = http.formvalue("device_id")
	local block = http.formvalue("block")
	if dev then
		os.execute('curl -s -X POST http://127.0.0.1:9999/api/block -d \'{"device_id":'..dev..',"block":'..(block == "1" and "true" or "false")..'}\' >/dev/null 2>&1')
	end
	http.prepare_content("application/json")
	http.write('{"ok":true}')
end

function api_limit()
	local http = require "luci.http"
	local dev = http.formvalue("device_id")
	local limit = http.formvalue("rate_limit") or "0"
	if dev then
		os.execute('curl -s -X POST http://127.0.0.1:9999/api/limit -d \'{"device_id":'..dev..',"rate_limit":'..limit..'}\' >/dev/null 2>&1')
	end
	http.prepare_content("application/json")
	http.write('{"ok":true}')
end
