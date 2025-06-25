import ReactMarkdown from 'react-markdown';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { atomDark } from 'react-syntax-highlighter/dist/esm/styles/prism';

interface TuiMarkdownProps {
  content: string;
}

export function TuiMarkdown({ content }: TuiMarkdownProps) {
  return (
    <div className="font-mono text-xs text-gray-300 leading-tight">
      <ReactMarkdown
        components={{
          // Headers - simple text, no large sizes
          h1: ({ children }) => (
            <div className="text-white font-semibold mt-2 mb-1">
              {children}
            </div>
          ),
          h2: ({ children }) => (
            <div className="text-white font-medium mt-2 mb-1">
              ## {children}
            </div>
          ),
          h3: ({ children }) => (
            <div className="text-gray-200 mt-1 mb-1">
              ### {children}
            </div>
          ),
          h4: ({ children }) => (
            <div className="text-gray-300 mt-1 mb-1">
              #### {children}
            </div>
          ),
          h5: ({ children }) => (
            <div className="text-gray-400 mt-1 mb-1">
              ##### {children}
            </div>
          ),
          h6: ({ children }) => (
            <div className="text-gray-400 mt-1 mb-1">
              ###### {children}
            </div>
          ),
          
          // Paragraphs - tight spacing
          p: ({ children }) => (
            <div className="mb-1">
              {children}
            </div>
          ),
          
          // Lists - render as plain text, no bullets
          ul: ({ children }) => (
            <div className="mb-1">
              {children}
            </div>
          ),
          ol: ({ children }) => (
            <div className="mb-1">
              {children}
            </div>
          ),
          li: ({ children, ordered, index }) => (
            <div className="text-gray-300">
              {ordered ? `${(index || 0) + 1}. ` : '- '}
              {children}
            </div>
          ),
          
          // Code blocks - subtle highlighting
          code: ({ inline, className, children, ...props }) => {
            const match = /language-(\w+)/.exec(className || '');
            return !inline && match ? (
              <SyntaxHighlighter
                style={atomDark}
                language={match[1]}
                PreTag="div"
                className="text-xs rounded-none border-none"
                customStyle={{
                  background: '#1a1a1a',
                  margin: '4px 0',
                  padding: '8px',
                  fontSize: '11px',
                  lineHeight: '1.2',
                }}
                {...props}
              >
                {String(children).replace(/\n$/, '')}
              </SyntaxHighlighter>
            ) : (
              <code className="bg-gray-800 text-gray-200 px-1 rounded-none text-xs" {...props}>
                {children}
              </code>
            );
          },
          
          // Links - subtle styling
          a: ({ children, href }) => (
            <a 
              href={href} 
              className="text-blue-400 underline hover:text-blue-300"
              target="_blank"
              rel="noopener noreferrer"
            >
              {children}
            </a>
          ),
          
          // Emphasis
          em: ({ children }) => (
            <em className="text-gray-200 italic">
              {children}
            </em>
          ),
          strong: ({ children }) => (
            <strong className="text-white font-semibold">
              {children}
            </strong>
          ),
          
          // Blockquotes
          blockquote: ({ children }) => (
            <div className="border-l-2 border-gray-600 pl-2 ml-1 text-gray-400 mb-1">
              {children}
            </div>
          ),
          
          // Horizontal rules
          hr: () => (
            <div className="border-t border-gray-700 my-2" />
          ),
          
          // Tables - simple styling
          table: ({ children }) => (
            <table className="border-collapse mb-2">
              {children}
            </table>
          ),
          th: ({ children }) => (
            <th className="border border-gray-600 px-2 py-1 text-left text-gray-200 text-xs">
              {children}
            </th>
          ),
          td: ({ children }) => (
            <td className="border border-gray-600 px-2 py-1 text-gray-300 text-xs">
              {children}
            </td>
          ),
        }}
      >
        {content}
      </ReactMarkdown>
    </div>
  );
}