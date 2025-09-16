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
  const [config, setConfig] = useState<{regions:string[]; allowMega:boolean; allowPrimal:boolean} | null>(null);

  const startWithConfig = async (c:{regions:string[]; allowMega:boolean; allowPrimal:boolean}) => {
    // call backend start directly
    const res = await fetch('/api/quiz/start', {method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({regions:c.regions, allowMega:c.allowMega, allowPrimal:c.allowPrimal})});
    const data = await res.json();
    setSession({sessionId:data.sessionId});
    setSeed(Date.now());
    setView('quiz');
  };

  return (
    <div style={{maxWidth: 960, margin: '0 auto', fontFamily:'system-ui'}}>
      {view === 'start' && (
        <StartScreen
          initialConfig={config}
          onStarted={(sessionId, c)=>{ setConfig(c); setSession({sessionId}); setView('quiz'); setSeed(Date.now()); }}
        />
      )}
      {view === 'quiz' && session && (
        <QuizScreen key={seed} session={session} onSolved={(p)=>{setSession({...session, ...p}); setView('result');}} onGiveUp={(p)=>{setSession({...session, ...p}); setView('result');}} onAbort={()=>setView('start')} />
      )}
      {view === 'result' && session && (
        <ResultScreen session={session} onNext={()=>{ if(config){ startWithConfig(config); } else { setView('start'); } }} onBack={()=>setView('start')} />
      )}
    </div>
  );
};
