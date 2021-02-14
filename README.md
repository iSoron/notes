# Notes

**Notes** is minimalist and collaborative wiki webserver with LaTeX support. Originally based on [cowyo](https://github.com/schollz/cowyo).

## Features

- Simplified user interface 
- Support for checkmarks and nested lists (similar to GitHub)
- Support for formulas with [MathJax](https://www.mathjax.org/)
- Support for diagrams with [Mermaid](https://mermaid-js.github.io/mermaid/#/)
- Live Markdown preview while editing

## Usage

```bash
docker run \
    --publish 8050:8050 \
    --volume $(pwd)/data:/data \
    isoron/notes
```

## License

```text
Copyright (c) 2020-2021 Alinson S. Xavier <git@axavier.org>
Copyright (c) 2016-2018 Zack Scholl

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```
