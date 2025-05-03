local target = "pico2-w"
local ok, info, json

ok, info = pcall(vim.fn.system, ("tinygo info -json %s"):format(target))
if not ok then
	vim.print("error calling tinygo: " .. info)
	return {}
end

ok, json = pcall(vim.fn.json_decode, info)
if not ok then
	vim.print("error decoding JSON output: " .. json)
	return {}
end

if not vim.fn.has_key(json, "goroot") or not vim.fn.has_key(json, "build_tags") then
	vim.print("unexpected JSON format")
	return {}
end

local goroot = json["goroot"]
local goflags = "-tags=" .. vim.fn.join(json["build_tags"], ",")

return {
	{
		"neovim/nvim-lspconfig",
		opts = {
			servers = {
				gopls = {
					cmd_env = {
						GOROOT = goroot,
						GOFLAGS = goflags,
					},
				},
			},
		},
	},
}
