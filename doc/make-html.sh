#!/usr/bin/env bash
#
# Generates HTML from markdown.
# Depends on the blackfriday markdown parser.
#

# Go: https://github.com/russross/blackfriday, https://github.com/russross/blackfriday-tool
BLACKFRIDAY="`type -P blackfriday-tool`"
TITLE="The Speak Interface Definition Language"
PREFIX="speak-spec"

cat > ${PREFIX}.html <<EOF
<!DOCTYPE HTML>
<html lang="en-US">
<head>
<meta charset="UTF-8">
<title>${TITLE}</title>
<link rel="stylesheet" href="style.css">
</head>
<body>
EOF

"$BLACKFRIDAY" ${PREFIX}.md >> ${PREFIX}.html

cat >> ${PREFIX}.html <<EOF
</body>
</html>
EOF
