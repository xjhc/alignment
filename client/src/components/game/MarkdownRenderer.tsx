import React from 'react';

interface MarkdownRendererProps {
  content: string;
  className?: string;
}

export const MarkdownRenderer: React.FC<MarkdownRendererProps> = ({ content, className = '' }) => {
  const parseMarkdown = (text: string): React.ReactNode[] => {
    const elements: React.ReactNode[] = [];
    
    // Handle line breaks and lists first
    const lines = text.split('\n');
    
    lines.forEach((line, lineIndex) => {
      if (lineIndex > 0) {
        elements.push(<br key={`br-${lineIndex}`} />);
      }
      
      // Check if line is a bullet point
      const listMatch = line.match(/^\s*\*\s+(.+)$/);
      if (listMatch) {
        elements.push(
          <div key={`list-${lineIndex}`} className="flex items-start gap-2 my-1">
            <span className="text-text-muted">•</span>
            <span>{parseInlineMarkdown(listMatch[1])}</span>
          </div>
        );
        return;
      }
      
      // Parse inline markdown for regular lines
      elements.push(...parseInlineMarkdown(line));
    });
    
    return elements;
  };
  
  const parseInlineMarkdown = (text: string): React.ReactNode[] => {
    const elements: React.ReactNode[] = [];
    let remaining = text;
    let keyCounter = 0;
    
    while (remaining.length > 0) {
      // Bold text: **text**
      const boldMatch = remaining.match(/^\*\*([^*]+)\*\*/);
      if (boldMatch) {
        elements.push(
          <strong key={`bold-${keyCounter++}`} className="font-bold text-text-primary">
            {boldMatch[1]}
          </strong>
        );
        remaining = remaining.slice(boldMatch[0].length);
        continue;
      }
      
      // Italic text: *text*
      const italicMatch = remaining.match(/^\*([^*]+)\*/);
      if (italicMatch) {
        elements.push(
          <em key={`italic-${keyCounter++}`} className="italic text-text-secondary">
            {italicMatch[1]}
          </em>
        );
        remaining = remaining.slice(italicMatch[0].length);
        continue;
      }
      
      // Strikethrough text: ~text~
      const strikeMatch = remaining.match(/^~([^~]+)~/);
      if (strikeMatch) {
        elements.push(
          <span key={`strike-${keyCounter++}`} className="line-through text-text-muted">
            {strikeMatch[1]}
          </span>
        );
        remaining = remaining.slice(strikeMatch[0].length);
        continue;
      }
      
      // Inline code: `text`
      const codeMatch = remaining.match(/^`([^`]+)`/);
      if (codeMatch) {
        elements.push(
          <code key={`code-${keyCounter++}`} className="bg-background-secondary text-text-primary px-1 py-0.5 rounded text-xs font-mono border border-border">
            {codeMatch[1]}
          </code>
        );
        remaining = remaining.slice(codeMatch[0].length);
        continue;
      }
      
      // @mentions: @PlayerName
      const mentionMatch = remaining.match(/^@(\w+)/);
      if (mentionMatch) {
        elements.push(
          <span key={`mention-${keyCounter++}`} className="bg-blue-500/20 text-blue-400 px-1 py-0.5 rounded font-medium">
            @{mentionMatch[1]}
          </span>
        );
        remaining = remaining.slice(mentionMatch[0].length);
        continue;
      }
      
      // Regular character
      const nextSpecialChar = remaining.search(/[\*~`@]/);
      if (nextSpecialChar === -1) {
        // No more special characters, add rest as text
        elements.push(remaining);
        break;
      } else if (nextSpecialChar > 0) {
        // Add text before next special character
        elements.push(remaining.slice(0, nextSpecialChar));
        remaining = remaining.slice(nextSpecialChar);
      } else {
        // Special character at start but no match, add it as regular text
        elements.push(remaining[0]);
        remaining = remaining.slice(1);
      }
    }
    
    return elements;
  };
  
  return (
    <div className={className}>
      {parseMarkdown(content)}
    </div>
  );
};