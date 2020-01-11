#! /usr/bin/env python3


what_is_embly = "../../app/blog/content/what-is-embly.md"
f = open(what_is_embly)

out = ""

extension_lang_map = dict(toml="toml", rs="rust", hcl="hcl")

in_code_block = False
for line in f.readlines():
    if "<!-- begin" in line:
        out += line
        in_code_block = True
        filename = line.split(" ")[2]
        syntax = extension_lang_map[filename.split(".")[-1]]
        code_file = open(filename).read()
        out += f"```{syntax}\n"
        out += code_file
        out += f"```\n"

    if "<!-- end" in line:
        in_code_block = False

    if not in_code_block:
        out += line

f.close()

with open(what_is_embly, "r+") as f:
    f.seek(0)
    f.write(out)
