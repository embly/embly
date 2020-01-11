#!/usr/bin/env python3

what_is_embly_html = "../../app/blog/dist/what-is-embly/index.html"


content = {}

with open(what_is_embly_html) as f:
    this_content = ""
    in_content = False
    for line in f.readlines():
        if "<!-- begin" in line:
            in_content = True

        if in_content:
            this_content += line

        if "<!-- end" in line:
            filename = line.split(" ")[2]
            in_content = False
            content[filename] = this_content
            this_content = ""


with open("../../app/blog/templates/index.html", "r+") as f:
    out = ""
    in_content = False
    for line in f.readlines():
        if "<!-- begin" in line:
            in_content = True
            filename = line.strip().split(" ")[2]
            out += content[filename]

        if not in_content:
            out += line

        if "<!-- end" in line:
            in_content = False
    f.seek(0)
    f.write(out)
