import type { ClaudeTodoWriteData } from '../types/api';

interface TodoWriteRendererProps {
  data: ClaudeTodoWriteData;
}

export function TodoWriteRenderer({ data }: TodoWriteRendererProps) {
  return (
    <div className="font-mono text-xs text-gray-300 leading-tight">
      <div className="text-orange-500 font-medium mb-1">● Update Todos</div>
      <div className="ml-2">
        {data.todos.map((todo) => {
          const checkbox = todo.status === 'completed' ? '☑' : 
                          todo.status === 'in_progress' ? '◐' : '☐';
          
          return (
            <div key={todo.id} className="text-gray-300 mb-0.5">
              <span className="text-blue-400">{checkbox}</span>
              <span className="ml-1">{todo.content}</span>
              {todo.priority === 'high' && (
                <span className="ml-2 text-red-400 text-xs">(high)</span>
              )}
              {todo.priority === 'medium' && (
                <span className="ml-2 text-yellow-400 text-xs">(medium)</span>
              )}
              {todo.priority === 'low' && (
                <span className="ml-2 text-gray-500 text-xs">(low)</span>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}