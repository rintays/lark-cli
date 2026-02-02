# Docs (Docx) Workflows

Docs in this CLI refer to Docx documents. Most commands accept a Docx token or a Docx URL.

## Read doc content

Markdown (default):

```bash
lark docs get <DOCX_TOKEN> --format md
```

Plain text:

```bash
lark docs get <DOCX_TOKEN> --format txt
```

Blocks (structured JSON/table):

```bash
lark docs get <DOCX_TOKEN> --format blocks --json
```

## Download Markdown then overwrite

Download to a file:

```bash
lark docs get <DOCX_TOKEN> --format md > doc.md
```

Edit `doc.md`, then overwrite the Docx content:

```bash
lark docs overwrite <DOCX_TOKEN> --content-file doc.md
```

## Convert Markdown/HTML to blocks

```bash
lark docs convert --content-type markdown --content "# Title"
```

## Overwrite with HTML

```bash
lark docs overwrite <DOCX_TOKEN> --content-type html --content-file page.html
```
