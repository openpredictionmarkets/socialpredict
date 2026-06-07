import React from 'react';

const inlineTokenPattern = /(\*\*[^*]+\*\*|\*[^*]+\*|`[^`]+`|\[[^\]]+\]\(https?:\/\/[^\s)]+\))/g;

const renderInlineToken = (token, key) => {
  if (token.startsWith('**') && token.endsWith('**')) {
    return <strong key={key}>{token.slice(2, -2)}</strong>;
  }
  if (token.startsWith('*') && token.endsWith('*')) {
    return <em key={key}>{token.slice(1, -1)}</em>;
  }
  if (token.startsWith('`') && token.endsWith('`')) {
    return <code key={key} className="rounded bg-gray-900 px-1 py-0.5 text-xs text-sky-100">{token.slice(1, -1)}</code>;
  }
  const linkMatch = token.match(/^\[([^\]]+)\]\((https?:\/\/[^\s)]+)\)$/);
  if (linkMatch) {
    return (
      <a key={key} href={linkMatch[2]} target="_blank" rel="noreferrer" className="text-sky-300 underline decoration-sky-400/50 hover:text-sky-200">
        {linkMatch[1]}
      </a>
    );
  }
  return token;
};

const renderInline = (text) => {
  const pieces = String(text || '').split(inlineTokenPattern).filter((piece) => piece !== '');
  return pieces.map((piece, index) => renderInlineToken(piece, `${index}-${piece}`));
};

const MarkdownLite = ({ children, className = '' }) => {
  const lines = String(children || '').split(/\r?\n/);
  const blocks = [];
  let listItems = [];
  let orderedList = false;

  const flushList = () => {
    if (!listItems.length) return;
    const ListTag = orderedList ? 'ol' : 'ul';
    blocks.push(
      <ListTag key={`list-${blocks.length}`} className={`${orderedList ? 'list-decimal' : 'list-disc'} space-y-1 pl-5`}>
        {listItems.map((item, index) => <li key={`${index}-${item}`}>{renderInline(item)}</li>)}
      </ListTag>,
    );
    listItems = [];
    orderedList = false;
  };

  lines.forEach((rawLine, index) => {
    const line = rawLine.trimEnd();
    const unordered = line.match(/^[-*]\s+(.+)$/);
    const ordered = line.match(/^\d+\.\s+(.+)$/);
    if (unordered || ordered) {
      const nextOrdered = Boolean(ordered);
      if (listItems.length && orderedList !== nextOrdered) {
        flushList();
      }
      orderedList = nextOrdered;
      listItems.push((ordered || unordered)[1]);
      return;
    }

    flushList();
    if (!line.trim()) {
      blocks.push(<div key={`space-${index}`} className="h-2" />);
      return;
    }
    if (line.trimStart().startsWith('>')) {
      blocks.push(
        <blockquote key={`quote-${index}`} className="border-l-2 border-sky-500/50 pl-3 text-gray-300">
          {renderInline(line.replace(/^\s*>\s?/, ''))}
        </blockquote>,
      );
      return;
    }
    blocks.push(<p key={`p-${index}`}>{renderInline(line)}</p>);
  });
  flushList();

  return <div className={`space-y-2 text-sm leading-6 ${className}`.trim()}>{blocks}</div>;
};

export default MarkdownLite;
