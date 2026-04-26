import re


def split_text(text: str, max_chars: int = 350) -> list[str]:
    text = " ".join(text.split())
    if not text:
        return []

    sentences = re.split(r"(?<=[.!?])\s+", text)
    chunks: list[str] = []
    current = ""
    for sentence in sentences:
        sentence = sentence.strip()
        if not sentence:
            continue
        if len(sentence) > max_chars:
            if current:
                chunks.append(current)
                current = ""
            chunks.extend(_split_long_sentence(sentence, max_chars))
            continue
        next_value = sentence if not current else f"{current} {sentence}"
        if len(next_value) <= max_chars:
            current = next_value
        else:
            chunks.append(current)
            current = sentence
    if current:
        chunks.append(current)
    return chunks


def _split_long_sentence(sentence: str, max_chars: int) -> list[str]:
    parts: list[str] = []
    current = ""
    for token in re.split(r"(\s+)", sentence):
        if len(current) + len(token) <= max_chars:
            current += token
            continue
        if current.strip():
            parts.append(current.strip())
        current = token.strip()
    if current.strip():
        parts.append(current.strip())
    return parts
