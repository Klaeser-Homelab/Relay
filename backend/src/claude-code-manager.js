import { query } from '@anthropic-ai/claude-code';

export class ClaudeCodeManager {
  constructor() {
    this.apiKey = process.env.ANTHROPIC_API_KEY;
    
    if (!this.apiKey) {
      console.warn('ANTHROPIC_API_KEY not provided. Claude Code functionality will be limited.');
    }
  }

  async createPlan(prompt, workingDirectory, onStreamMessage = null, claudeSessionId = null) {
    if (!this.apiKey) {
      throw new Error('ANTHROPIC_API_KEY is required for Claude Code functionality');
    }

    if (!prompt) {
      throw new Error('Prompt is required for planning');
    }

    console.log('=== CLAUDE CODE: Starting session ===');
    console.log('Prompt:', prompt);
    console.log('Working directory:', workingDirectory);

    try {
      const responses = [];
      let fullResponse = '';
      let todos = [];
      let sessionId = claudeSessionId; // Track session ID

      // Use Claude Code SDK with specified working directory
      const enhancedPrompt = claudeSessionId ? 
        prompt : // For continuation, use prompt as-is
        `${prompt}

IMPORTANT: Please break down your response into clear, actionable steps using the TodoWrite tool. This helps provide real-time feedback on your progress. Create todos for major steps in your analysis, planning, or implementation approach.

Use the TodoWrite tool to:
1. Show your thinking process
2. Break complex tasks into manageable steps  
3. Track progress through the implementation
4. Provide clear milestones

Provide comprehensive, detailed responses with specific examples and best practices.`;

      const queryOptions = {
        prompt: enhancedPrompt,
        options: {
          maxTurns: 10 // Allow multiple turns for complete interaction
        }
      };
      
      // Add session continuation if we have a previous session
      if (claudeSessionId) {
        console.log(`✓ Continuing conversation with session ID: ${claudeSessionId}`);
        queryOptions.options.resume = claudeSessionId;
      }

      // Set working directory if provided
      if (workingDirectory) {
        queryOptions.cwd = workingDirectory;
      }

      console.log('✓ Starting Claude Code query with options:', JSON.stringify(queryOptions, null, 2));

      // Stream responses from Claude Code (streaming JSON format)
      for await (const message of query(queryOptions)) {
        console.log('✓ Received Claude Code message:', JSON.stringify(message, null, 2));
        
        // Capture session ID for conversation continuation
        if (message.session_id && !sessionId) {
          sessionId = message.session_id;
          console.log(`✓ Captured session ID: ${sessionId}`);
        }
        
        if (message.type === 'text') {
          // Handle direct text responses
          console.log('✓ Text content received:', message.content);
          fullResponse += message.content;
          responses.push({
            type: 'text',
            content: message.content,
            timestamp: new Date().toISOString()
          });
          
          // Stream text content immediately
          if (onStreamMessage) {
            onStreamMessage({
              type: 'claude_text',
              content: message.content,
              timestamp: new Date().toISOString()
            });
          }
        } else if (message.type === 'assistant' && message.message?.content) {
          // Handle assistant messages with content array
          const assistantContent = message.message.content
            .filter(item => item.type === 'text')
            .map(item => item.text)
            .join('');
          
          // Stream text content immediately
          if (assistantContent && onStreamMessage) {
            console.log('✓ Streaming assistant text:', assistantContent.substring(0, 100) + '...');
            onStreamMessage({
              type: 'claude_text',
              content: assistantContent,
              timestamp: new Date().toISOString()
            });
          }
          
          // Extract and stream todo tool calls
          const toolUses = message.message.content.filter(item => item.type === 'tool_use');
          for (const toolUse of toolUses) {
            if (toolUse.name === 'TodoWrite' && toolUse.input?.todos) {
              console.log('✓ Found todos in tool use:', toolUse.input.todos.length, 'items');
              todos = todos.concat(toolUse.input.todos);
              
              // Stream TodoWrite tool call immediately
              if (onStreamMessage) {
                console.log('✓ Streaming TodoWrite tool call');
                onStreamMessage({
                  type: 'claude_todowrite',
                  content: {
                    todos: toolUse.input.todos
                  },
                  timestamp: new Date().toISOString()
                });
              }
            }
          }
          
          if (assistantContent) {
            console.log('✓ Assistant content received:', assistantContent.substring(0, 100) + '...');
            fullResponse += assistantContent;
            responses.push({
              type: 'assistant',
              content: assistantContent,
              timestamp: new Date().toISOString()
            });
          }
        } else if (message.type === 'result' && message.result) {
          // Handle final result message - this is just a summary, don't override fullResponse
          console.log('✓ Final result received:', message.result.substring(0, 100) + '...');
          responses.push({
            type: 'result',
            content: message.result,
            timestamp: new Date().toISOString()
          });
        } else if (message.type === 'error') {
          console.error('❌ Claude Code error:', message.content || message.error);
          throw new Error(`Claude Code error: ${message.content || message.error}`);
        } else {
          console.log('✓ Other message type:', message.type);
          
          // Check for TodoWrite tool calls in any other message types
          if (message.content || message.message) {
            const messageContent = message.content || message.message;
            if (Array.isArray(messageContent)) {
              const toolUses = messageContent.filter(item => item.type === 'tool_use');
              for (const toolUse of toolUses) {
                if (toolUse.name === 'TodoWrite' && toolUse.input?.todos) {
                  console.log('✓ Found todos in other message type:', toolUse.input.todos.length, 'items');
                  todos = todos.concat(toolUse.input.todos);
                  
                  // Stream TodoWrite tool call immediately
                  if (onStreamMessage) {
                    console.log('✓ Streaming TodoWrite from other message type');
                    onStreamMessage({
                      type: 'claude_todowrite',
                      content: {
                        todos: toolUse.input.todos
                      },
                      timestamp: new Date().toISOString()
                    });
                  }
                }
              }
            }
          }
        }
      }

      // Format the complete plan with todos if available
      let formattedPlan = fullResponse;
      let todosSection = '';
      
      if (todos.length > 0) {
        // Deduplicate todos by id, keeping the latest version
        const uniqueTodos = {};
        todos.forEach(todo => {
          uniqueTodos[todo.id] = todo;
        });
        const finalTodos = Object.values(uniqueTodos);
        
        // Format todos with proper status indicators
        todosSection = '\n\n# Todos\n\n' + 
          finalTodos.map(todo => {
            const checkbox = todo.status === 'completed' ? '[x]' : 
                           todo.status === 'in_progress' ? '[.]' : '[ ]';
            return `${checkbox} ${todo.content} (${todo.priority} priority)`;
          }).join('\n');
        
        formattedPlan = fullResponse + todosSection;
        console.log('✓ Added', finalTodos.length, 'unique todos to plan');
      }

      console.log('✓ Claude Code session completed');
      console.log('Total response length:', formattedPlan.length);

      return {
        success: true,
        plan: formattedPlan,
        responses,
        todos,
        workingDirectory,
        sessionId, // Include session ID for continuation
        timestamp: new Date().toISOString()
      };

    } catch (error) {
      console.error('❌ Claude Code planning failed:', error);
      throw new Error(`Failed to create plan: ${error.message}`);
    }
  }

  async streamPlan(prompt, workingDirectory, onMessage) {
    if (!this.apiKey) {
      throw new Error('ANTHROPIC_API_KEY is required for Claude Code functionality');
    }

    if (!prompt) {
      throw new Error('Prompt is required for planning');
    }

    if (typeof onMessage !== 'function') {
      throw new Error('onMessage callback is required for streaming');
    }

    console.log('=== CLAUDE CODE: Starting streaming planning session ===');
    console.log('Prompt:', prompt);
    console.log('Working directory:', workingDirectory);

    try {
      const queryOptions = {
        prompt,
        options: {
          maxTurns: 1
        }
      };

      if (workingDirectory) {
        queryOptions.cwd = workingDirectory;
      }

      console.log('✓ Starting streaming Claude Code query');

      for await (const message of query(queryOptions)) {
        console.log('✓ Streaming message:', message.type);
        
        // Call the callback with each message
        onMessage({
          type: message.type,
          content: message.content,
          timestamp: new Date().toISOString()
        });

        if (message.type === 'error') {
          throw new Error(`Claude Code error: ${message.content}`);
        }
      }

      console.log('✓ Streaming Claude Code session completed');
      
      // Signal completion
      onMessage({
        type: 'complete',
        content: null,
        timestamp: new Date().toISOString()
      });

    } catch (error) {
      console.error('❌ Streaming Claude Code planning failed:', error);
      
      // Signal error to callback
      onMessage({
        type: 'error',
        content: error.message,
        timestamp: new Date().toISOString()
      });
      
      throw error;
    }
  }
}