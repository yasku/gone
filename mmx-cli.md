# mmx-cli

MiniMax CLI for text, image, video, speech, music, vision and search.

## Install
```bash
npm install -g mmx-cli
mmx auth login --api-key YOUR_API_KEY
```

## Commands

| Command | Description |
|---------|-------------|
| `mmx search query "text"` | Web search |
| `mmx vision /path/image "question"` | Analyze image |
| `mmx text chat --message "text"` | Text interaction |
| `mmx image "description"` | Generate image |
| `mmx video generate --prompt "description"` | Generate video |
| `mmx music generate --prompt "description" --out file.mp3` | Generate music |
| `mmx speech synthesize --text "text" --out file.mp3` | Text-to-speech |

## Usage Examples

```bash
# Search
mmx search query "golang bubble tea tui framework"

# Vision (analyze screenshot you have)
mmx vision ~/Downloads/screenshot.png "What app is this?"

# Text
mmx text chat --message "Explain Go channel patterns"

# Generate image
mmx image "cyberpunk city night scene 16:9"
```

## Docs
https://platform.minimax.io/docs/token-plan/minimax-cli