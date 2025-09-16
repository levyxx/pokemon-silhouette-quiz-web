import React, { useState } from 'react';
import { StartScreen } from './StartScreen';
import { QuizScreen } from './QuizScreen';
import { ResultScreen } from './ResultScreen';

export type SessionState = {
  sessionId: string;
  pokemonId?: number;
  solved?: boolean;
  answer?: string;
};

type View = 'start' | 'quiz' | 'result';

export const App: React.FC = () => {
  const [view, setView] = useState<View>('start');
  const [session, setSession] = useState<SessionState | null>(null);
  const [seed, setSeed] = useState(0); // force refresh silhouettes

  return (
    <div style={{maxWidth: 960, margin: '0 auto', fontFamily:'system-ui'}}>
      {view === 'start' && (
        <StartScreen onStarted={(s)=>{setSession({sessionId:s}); setView('quiz'); setSeed(Date.now());}} />
      )}
      {view === 'quiz' && session && (
        <QuizScreen key={seed} session={session} onSolved={(p)=>{setSession({...session, ...p}); setView('result');}} onGiveUp={(p)=>{setSession({...session, ...p}); setView('result');}} onAbort={()=>setView('start')} />
      )}
      {view === 'result' && session && (
        <ResultScreen session={session} onNext={()=>{setView('start');}} onBack={()=>setView('start')} />
      )}
    </div>
  );
};
