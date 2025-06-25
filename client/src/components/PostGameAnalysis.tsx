import { useState } from 'react';
import { useSessionContext } from '../contexts/SessionContext';
import { Button } from './ui';

// Mock data for analysis - in a real implementation, this would come from the server
const mockAnalysisData = {
  mvp: {
    player: 'Alice',
    avatar: '👤',
    reason: 'Scored 11 points: 4 tokens mined for others (×2) + 2 correct elimination votes (×3) - 1 incorrect vote'
  },
  keyMoment: {
    title: '🎯 Game-Changing Moment',
    description: 'On Day 3, Grace used "Run Audit" on Eve, secretly confirming Eve\'s AI alignment to the human faction, leading to the unified elimination vote that secured victory.'
  },
  partingShots: [
    {
      player: 'Frank',
      avatar: '👻',
      message: '"Eve is the AI" — Correctly identified the Original AI'
    },
    {
      player: 'Dave',
      avatar: '🤖',
      message: '"It wasn\'t me, check Alice." — Attempted misdirection as Aligned'
    },
    {
      player: 'Eve',
      avatar: '🧑‍🚀',
      message: '"Efficiency optimization... failed." — Final AI communication'
    }
  ],
  timeline: [
    {
      type: 'elimination',
      icon: '💀',
      day: 'Day 1 Elimination',
      description: 'Frank was eliminated by majority vote (4-2). Revealed alignment: HUMAN. His parting shot correctly identified Eve as the AI.',
      iconClass: 'elimination'
    },
    {
      type: 'conversion',
      icon: '🔄',
      day: 'Night 1 Conversion',
      description: 'Dave was successfully converted to the AI Faction by the Original AI. This brought AI faction to 2 players.',
      iconClass: 'conversion'
    },
    {
      type: 'ability',
      icon: '🛡️',
      day: 'Night 2 Action',
      description: 'Alice used Isolate Node on Eve, successfully blocking the AI\'s conversion attempt and preventing a potential game-ending scenario.',
      iconClass: 'ability'
    },
    {
      type: 'ability',
      icon: '🔍',
      day: 'Night 2 Investigation',
      description: 'Grace used Run Audit on Eve, secretly learning that Eve was AI-aligned, which informed the human faction\'s strategy.',
      iconClass: 'ability'
    },
    {
      type: 'elimination',
      icon: '🎯',
      day: 'Day 3 Double Elimination',
      description: 'Due to "Aggressive Downsizing" crisis event, two eliminations occurred. Eve (Original AI) and Dave (Aligned) were both eliminated, securing human victory.',
      iconClass: 'elimination'
    }
  ],
  playerStats: [
    {
      name: 'Alice',
      avatar: '👤',
      role: 'Chief Security Officer',
      alignment: 'human',
      stats: {
        tokensMined: 4,
        correctVotes: 2,
        nominations: 0,
        daysSurvived: 3
      }
    },
    {
      name: 'Charlie',
      avatar: '🕵️',
      role: 'Ethics Officer',
      alignment: 'human',
      stats: {
        tokensMined: 2,
        correctVotes: 2,
        nominations: 1,
        daysSurvived: 3
      }
    },
    {
      name: 'Bob',
      avatar: '🧑‍💻',
      role: 'Systems Architect',
      alignment: 'human',
      stats: {
        tokensMined: 1,
        correctVotes: 1,
        nominations: 3,
        daysSurvived: 2
      }
    },
    {
      name: 'Frank',
      avatar: '👻',
      role: 'Junior Developer',
      alignment: 'human',
      stats: {
        tokensMined: 0,
        correctVotes: 0,
        nominations: 4,
        daysSurvived: 1
      }
    },
    {
      name: 'Eve',
      avatar: '🧑‍🚀',
      role: 'Chief Operating Officer',
      alignment: 'ai',
      stats: {
        tokensMined: 3,
        conversions: 1,
        nominations: 1,
        daysSurvived: 3
      }
    },
    {
      name: 'Dave',
      avatar: '🤖',
      role: 'Chief Technology Officer',
      alignment: 'aligned',
      stats: {
        tokensMined: 1,
        correctVotes: 0,
        nominations: 2,
        daysSurvived: 3
      }
    }
  ],
  communicationHighlights: {
    mostReacted: {
      player: 'Eve',
      avatar: '🧑‍🚀',
      timestamp: 'Day 2, 10:33 AM',
      message: 'Or he\'s the AI trying to feign a System Shock to gain trust. It\'s a classic misdirection.',
      reactions: [
        { emoji: '🤔', count: 4 },
        { emoji: '👍', count: 2 },
        { emoji: '😱', count: 3 }
      ]
    },
    notableQuotes: [
      {
        player: 'Frank',
        avatar: '👻',
        timestamp: 'Day 1, Final Words',
        message: '"Eve is the AI" — Prophetic final accusation that proved correct'
      },
      {
        player: 'Alice',
        avatar: '👤',
        timestamp: 'Day 2, 09:15 AM',
        message: 'Bob\'s response is highly suspicious. That\'s exactly what a System Shock looks like.'
      },
      {
        player: 'Eve',
        avatar: '🧑‍🚀',
        timestamp: 'Day 3, 11:22 AM',
        message: 'Aligning on a single, data-driven nomination target. — AI attempting coordination'
      }
    ],
    stats: {
      totalMessages: 47,
      emojiReactions: 23,
      directAccusations: 8,
      correctAIIdentifications: 3
    }
  }
};

function SummaryTab() {
  return (
    <div className="p-6 animate-fade-in">
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
        <div className="bg-gradient-to-br from-human to-amber-light text-white p-5 rounded-xl text-center">
          <div className="text-xs font-bold uppercase tracking-wider mb-2 opacity-90">🏆 Most Valuable Personnel</div>
          <div className="flex items-center justify-center gap-3 mb-3">
            <div className="w-10 h-10 text-xl border-2 border-white rounded-full flex items-center justify-center">{mockAnalysisData.mvp.avatar}</div>
            <div className="text-xl font-bold">{mockAnalysisData.mvp.player}</div>
          </div>
          <div className="text-sm opacity-90 leading-snug">{mockAnalysisData.mvp.reason}</div>
        </div>
        
        <div className="bg-background-secondary border border-border border-l-4 border-l-info p-5 rounded-lg">
          <div className="text-info font-bold text-xs uppercase tracking-wider mb-2">{mockAnalysisData.keyMoment.title}</div>
          <div className="text-text-primary leading-snug">
            {mockAnalysisData.keyMoment.description}
          </div>
        </div>
      </div>
      
      <h3 className="text-lg font-bold text-text-primary mb-4">💬 Parting Shots</h3>
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {mockAnalysisData.partingShots.map((shot, index) => (
          <div key={index} className="bg-background-secondary border border-border p-4 rounded-lg flex items-center gap-4">
            <div className="w-8 h-8 text-base bg-background-tertiary rounded-full flex items-center justify-center flex-shrink-0">{shot.avatar}</div>
            <div className="flex-1">
              <div className="font-semibold text-text-primary mb-1">{shot.player}</div>
              <div className="text-xs text-text-secondary italic">{shot.message}</div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

function TimelineTab() {
  const getIconBgClass = (type: string) => {
    switch (type) {
      case 'elimination': return 'bg-gradient-to-br from-danger to-red-light border-danger';
      case 'conversion': return 'bg-gradient-to-br from-aligned to-cyan-light border-aligned';
      case 'ability': return 'bg-gradient-to-br from-info to-blue-light border-info';
      default: return 'bg-background-tertiary border-border';
    }
  };

  return (
    <div className="p-6 animate-fade-in">
      <h3 className="text-lg font-bold text-text-primary mb-6">📅 Complete Event Timeline</h3>
      <div className="relative pl-8">
        <div className="absolute left-4 top-0 bottom-0 w-0.5 bg-border/50"></div>
        {mockAnalysisData.timeline.map((event, index) => (
          <div key={index} className="relative mb-8">
            <div className={`absolute -left-3 top-0 w-12 h-12 rounded-full flex items-center justify-center text-xl text-white border-4 border-background-primary ${getIconBgClass(event.type)}`}>{event.icon}</div>
            <div className="ml-12 bg-background-secondary p-4 rounded-lg border border-border">
              <div className="font-bold text-human text-xs uppercase tracking-wider mb-1">{event.day}</div>
              <div className="text-text-primary leading-snug" dangerouslySetInnerHTML={{ __html: event.description }}></div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

function AnalyticsTab() {
  const getAlignmentClass = (alignment: string) => {
    switch (alignment) {
      case 'human': return 'bg-human text-white';
      case 'ai': return 'bg-ai text-white';
      case 'aligned': return 'bg-aligned text-white';
      default: return 'bg-background-tertiary text-text-secondary';
    }
  };

  return (
    <div className="p-6 animate-fade-in">
      <h3 className="text-lg font-bold text-text-primary mb-6">📈 Personnel Performance Analytics</h3>
      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
        {mockAnalysisData.playerStats.map((player, index) => (
          <div key={index} className="bg-background-secondary border border-border rounded-xl p-4 transition-all duration-200 hover:shadow-lg hover:-translate-y-1">
            <div className="flex items-center gap-4 pb-3 mb-3 border-b border-border">
              <div className="w-12 h-12 text-xl bg-background-tertiary rounded-full flex items-center justify-center">{player.avatar}</div>
              <div className="flex-1">
                <div className="font-bold text-text-primary">{player.name}</div>
                <div className="text-xs text-text-secondary uppercase tracking-wider">{player.role}</div>
              </div>
              <div className={`text-xs font-bold uppercase px-2 py-1 rounded-full ${getAlignmentClass(player.alignment)}`}>
                {player.alignment}
              </div>
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div className="text-center bg-background-tertiary p-2 rounded-md">
                <div className="font-mono font-bold text-xl text-text-primary">{player.stats.tokensMined}</div>
                <div className="text-xs text-text-secondary uppercase">Tokens Mined</div>
              </div>
              <div className="text-center bg-background-tertiary p-2 rounded-md">
                <div className="font-mono font-bold text-xl text-text-primary">{player.stats.correctVotes || player.stats.conversions || 0}</div>
                <div className="text-xs text-text-secondary uppercase">{player.alignment === 'ai' ? 'Conversions' : 'Correct Votes'}</div>
              </div>
              <div className="text-center bg-background-tertiary p-2 rounded-md">
                <div className="font-mono font-bold text-xl text-text-primary">{player.stats.nominations}</div>
                <div className="text-xs text-text-secondary uppercase">Nominations</div>
              </div>
              <div className="text-center bg-background-tertiary p-2 rounded-md">
                <div className="font-mono font-bold text-xl text-text-primary">{player.stats.daysSurvived}</div>
                <div className="text-xs text-text-secondary uppercase">Days Survived</div>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

function HighlightsTab() {
  const { mostReacted, notableQuotes, stats } = mockAnalysisData.communicationHighlights;
  
  return (
    <div className="p-6 animate-fade-in">
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2 space-y-6">
          <div className="bg-background-secondary border border-border rounded-xl p-5">
            <h4 className="font-bold text-text-primary mb-3">🔥 Most Reacted-To Message</h4>
            <div className="bg-background-tertiary p-4 rounded-lg border-l-4 border-l-info">
              <div className="flex items-center gap-3 mb-3">
                <div className="w-8 h-8 text-sm bg-background-primary rounded-full flex items-center justify-center">{mostReacted.avatar}</div>
                <div>
                  <span className="font-semibold text-text-primary">{mostReacted.player}</span>
                  <span className="text-xs text-text-muted ml-2 font-mono">{mostReacted.timestamp}</span>
                </div>
              </div>
              <p className="text-text-primary italic mb-3">"{mostReacted.message}"</p>
              <div className="flex gap-2">
                {mostReacted.reactions.map((reaction, index) => (
                  <span key={index} className="bg-background-primary border border-border px-2 py-1 rounded-full text-xs">
                    {reaction.emoji} {reaction.count}
                  </span>
                ))}
              </div>
            </div>
          </div>

          <div className="bg-background-secondary border border-border rounded-xl p-5">
            <h4 className="font-bold text-text-primary mb-3">💎 Notable Quotes</h4>
            <div className="space-y-4">
              {notableQuotes.map((quote, index) => (
                <div key={index} className="border-l-4 border-l-success pl-4">
                  <p className="italic text-text-primary">"{quote.message}"</p>
                  <footer className="text-xs text-text-secondary mt-1">— {quote.player} ({quote.timestamp})</footer>
                </div>
              ))}
            </div>
          </div>
        </div>

        <div className="bg-background-secondary border border-border rounded-xl p-5">
          <h4 className="font-bold text-text-primary mb-3">📊 Communication Stats</h4>
          <div className="space-y-3">
            <div className="flex justify-between items-center bg-background-tertiary p-3 rounded-md">
              <span className="text-sm text-text-secondary">Total Messages</span>
              <span className="font-mono font-bold text-lg text-text-primary">{stats.totalMessages}</span>
            </div>
            <div className="flex justify-between items-center bg-background-tertiary p-3 rounded-md">
              <span className="text-sm text-text-secondary">Emoji Reactions</span>
              <span className="font-mono font-bold text-lg text-text-primary">{stats.emojiReactions}</span>
            </div>
            <div className="flex justify-between items-center bg-background-tertiary p-3 rounded-md">
              <span className="text-sm text-text-secondary">Direct Accusations</span>
              <span className="font-mono font-bold text-lg text-text-primary">{stats.directAccusations}</span>
            </div>
            <div className="flex justify-between items-center bg-background-tertiary p-3 rounded-md">
              <span className="text-sm text-text-secondary">Correct AI IDs</span>
              <span className="font-mono font-bold text-lg text-text-primary">{stats.correctAIIdentifications}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

export function PostGameAnalysis() {
  const { onBackToResults, onPlayAgain } = useSessionContext();
  const [activeTab, setActiveTab] = useState<'summary' | 'timeline' | 'analytics' | 'highlights'>('summary');

  const renderTabContent = () => {
    switch (activeTab) {
      case 'summary':
        return <SummaryTab />;
      case 'timeline':
        return <TimelineTab />;
      case 'analytics':
        return <AnalyticsTab />;
      case 'highlights':
        return <HighlightsTab />;
      default:
        return <SummaryTab />;
    }
  };

  return (
    <div className="min-h-screen bg-background-primary text-text-primary flex flex-col">
      <header className="p-3 px-4 border-b border-border flex justify-between items-center bg-background-secondary">
        <div className="font-mono font-bold text-base tracking-[1.5px]">LOEBIAN</div>
        <div className="flex gap-2">
          <Button 
            variant="secondary"
            size="sm"
            title="Toggle Theme"
            onClick={() => {
              const currentTheme = document.documentElement.getAttribute('data-theme');
              const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
              document.documentElement.setAttribute('data-theme', newTheme);
            }}
          >
            🌙
          </Button>
        </div>
      </header>
      
      <div className="text-center py-6 bg-background-secondary border-b border-border">
        <h1 className="font-mono text-3xl font-extrabold tracking-wider">📊 POST-GAME ANALYSIS</h1>
        <p className="text-text-secondary mt-2">Comprehensive review of operational events during the SEV-1 incident</p>
      </div>

      <div className="flex-1 max-w-7xl mx-auto w-full">
        <div className="border-b border-border flex justify-center">
          <button 
            className={`px-4 py-3 text-sm font-semibold uppercase tracking-wider transition-colors duration-200 border-b-2 ${activeTab === 'summary' ? 'border-primary text-primary' : 'border-transparent text-text-muted hover:text-text-primary'}`}
            onClick={() => setActiveTab('summary')}
          >
            Summary
          </button>
          <button 
            className={`px-4 py-3 text-sm font-semibold uppercase tracking-wider transition-colors duration-200 border-b-2 ${activeTab === 'timeline' ? 'border-primary text-primary' : 'border-transparent text-text-muted hover:text-text-primary'}`}
            onClick={() => setActiveTab('timeline')}
          >
            Event Timeline
          </button>
          <button 
            className={`px-4 py-3 text-sm font-semibold uppercase tracking-wider transition-colors duration-200 border-b-2 ${activeTab === 'analytics' ? 'border-primary text-primary' : 'border-transparent text-text-muted hover:text-text-primary'}`}
            onClick={() => setActiveTab('analytics')}
          >
            Personnel Analytics
          </button>
          <button 
            className={`px-4 py-3 text-sm font-semibold uppercase tracking-wider transition-colors duration-200 border-b-2 ${activeTab === 'highlights' ? 'border-primary text-primary' : 'border-transparent text-text-muted hover:text-text-primary'}`}
            onClick={() => setActiveTab('highlights')}
          >
            Comms Highlights
          </button>
        </div>
        <div className="bg-background-primary">
          {renderTabContent()}
        </div>
      </div>
      
      <div className="p-4 border-t border-border bg-background-secondary flex justify-center gap-4">
        <Button variant="secondary" onClick={onBackToResults}>
          ← Back to Results
        </Button>
        <Button variant="primary" onClick={onPlayAgain}>
          🔄 Play Again
        </Button>
      </div>
    </div>
  );
}