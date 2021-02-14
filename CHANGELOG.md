# Changelog

## [0.1.0] - 2020-02-07
### Added
- Add support for checkmarks and nested lists (similar to GitHub)
- Allow user to add formulas with [MathJax](https://www.mathjax.org/)
- Allow user to create diagrams with [Mermaid](https://mermaid-js.github.io/mermaid/#/)
- When editing, show raw content and preview side-by-side

### Changed
- Better organize data directory
- Simplify navigation bar
- Switch Markdown engine from [blackfriday](https://github.com/russross/blackfriday) to [goldmark](https://github.com/yuin/goldmark)
- Always sanitize Markdown output with [bluemonday](https://github.com/microcosm-cc/bluemonday)
- Use random strings for URLs instead of animal names
- Page publishing now simply generates a read-only URL

### Removed
- Remove built-in TLS support (use a reverse proxy instead)
- Remove cowyo lists
- Remove custom CSS
- Remove multi-website wikis
- Remove page encryption
- Remove page locking
- Remove self-destructing pages
- Remove sitemap.xml
