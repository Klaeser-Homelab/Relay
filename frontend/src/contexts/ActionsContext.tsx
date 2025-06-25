import React, { createContext, useState, ReactNode, Dispatch, SetStateAction, useContext } from "react";

export type Action =
  | { type: 'send_text'; message: string }
  | { type: 'send_audio'; audio: string }
  | { type: 'plan'; plan: string }
  | { type: 'challenge'; challenge: string };

export interface ActionsContextType {
  currentAction: Action | null;
  setCurrentAction: Dispatch<SetStateAction<Action | null>>;
}

export const ActionsContext = createContext<ActionsContextType | undefined>(undefined);

interface ActionsProviderProps {
  children: ReactNode;
}

export const ActionsProvider: React.FC<ActionsProviderProps> = ({ children }) => {
  const [currentAction, setCurrentAction] = useState<Action | null>(null);


  const values = {
    currentAction,
    setCurrentAction
  }

  return (
    <ActionsContext.Provider value={values}>
      {children}
    </ActionsContext.Provider>
  );
};

export const useActions = () => {
  const context = useContext(ActionsContext);
  if (!context) {
    throw new Error('useActions must be used within an ActionsProvider');
  }
  return context;
};