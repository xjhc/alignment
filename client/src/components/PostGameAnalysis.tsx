import { useState } from 'react';
import { GameState } from '../types';

interface PostGameAnalysisProps {
  gameState: GameState;
  onBackToResults: () => void;
  onPlayAgain: () => void;
}


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
    <div className="hud-pane active">
      <div className="mvp-section">
        <div className="mvp-card">
          <div className="mvp-title">🏆 Most Valuable Personnel</div>
          <div className="mvp-player">
            <div className="player-avatar">{mockAnalysisData.mvp.avatar}</div>
            <div className="mvp-name">{mockAnalysisData.mvp.player}</div>
          </div>
          <div className="mvp-reason">
            {mockAnalysisData.mvp.reason}
          </div>
        </div>
        
        <div className="key-moment-card">
          <div className="moment-title">{mockAnalysisData.keyMoment.title}</div>
          <div className="moment-description">
            {mockAnalysisData.keyMoment.description}
          </div>
        </div>
      </div>
      
      <h3 className="hud-section-title">💬 Parting Shots</h3>
      <div className="parting-shots-grid">
        {mockAnalysisData.partingShots.map((shot, index) => (
          <div key={index} className="parting-shot-item">
            <div className="player-avatar">{shot.avatar}</div>
            <div className="parting-shot-content">
              <div className="parting-shot-name">{shot.player}</div>
              <div className="parting-shot-message">{shot.message}</div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

function TimelineTab() {
  return (
    <div className="hud-pane active">
      <h3 className="hud-section-title">📅 Complete Event Timeline</h3>
      <div className="analysis-timeline">
        {mockAnalysisData.timeline.map((event, index) => (
          <div key={index} className={`timeline-item ${event.iconClass}`}>
            <div className="timeline-icon">{event.icon}</div>
            <div className="timeline-content">
              <div className="day">{event.day}</div>
              <div className="desc" dangerouslySetInnerHTML={{ __html: event.description }}></div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

function AnalyticsTab() {
  return (
    <div className="hud-pane active">
      <h3 className="hud-section-title">📈 Personnel Performance Analytics</h3>
      <div className="analytics-grid">
        {mockAnalysisData.playerStats.map((player, index) => (
          <div key={index} className="analytics-card">
            <div className="analytics-card-header">
              <div className="player-avatar">{player.avatar}</div>
              <div className="analytics-player-info">
                <div className="analytics-player-name">{player.name}</div>
                <div className="analytics-player-role">{player.role}</div>
              </div>
              <div className={`analytics-player-alignment ${player.alignment}`}>
                {player.alignment === 'human' ? 'Human' : 
                 player.alignment === 'ai' ? 'Original AI' : 'Aligned'}
              </div>
            </div>
            <div className="analytics-stats">
              <div className="analytics-stat-item">
                <div className="analytics-stat-value">{player.stats.tokensMined}</div>
                <div className="analytics-stat-label">Tokens Mined</div>
              </div>
              <div className="analytics-stat-item">
                <div className="analytics-stat-value">
                  {player.stats.correctVotes || player.stats.conversions || 0}
                </div>
                <div className="analytics-stat-label">
                  {player.alignment === 'ai' ? 'Conversion' : 'Correct Votes'}
                </div>
              </div>
              <div className="analytics-stat-item">
                <div className="analytics-stat-value">{player.stats.nominations}</div>
                <div className="analytics-stat-label">Nominations</div>
              </div>
              <div className="analytics-stat-item">
                <div className="analytics-stat-value">{player.stats.daysSurvived}</div>
                <div className="analytics-stat-label">Days Survived</div>
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
    <div className="hud-pane active">
      <div className="highlights-section">
        <div className="highlight-card">
          <div className="highlight-title">🔥 Most Reacted-To Message</div>
          <div className="message-highlight">
            <div className="message-meta">
              <div className="player-avatar">{mostReacted.avatar}</div>
              <span className="message-author-name">{mostReacted.player}</span>
              <span className="message-timestamp">{mostReacted.timestamp}</span>
            </div>
            <div className="message-text">
              {mostReacted.message}
            </div>
            <div className="message-reactions-display">
              {mostReacted.reactions.map((reaction, index) => (
                <span key={index} className="reaction-display">
                  {reaction.emoji} {reaction.count}
                </span>
              ))}
            </div>
          </div>
        </div>
        
        <div className="highlight-card">
          <div className="highlight-title">💎 Notable Quotes</div>
          <div className="quote-grid">
            {notableQuotes.map((quote, index) => (
              <div key={index} className="quote-item">
                <div className="message-meta">
                  <div className="player-avatar">{quote.avatar}</div>
                  <span className="message-author-name">{quote.player}</span>
                  <span className="message-timestamp">{quote.timestamp}</span>
                </div>
                <div className="message-text">{quote.message}</div>
              </div>
            ))}
          </div>
        </div>
        
        <div className="highlight-card">
          <div className="highlight-title">📊 Communication Stats</div>
          <div className="analytics-stats">
            <div className="analytics-stat-item">
              <div className="analytics-stat-value">{stats.totalMessages}</div>
              <div className="analytics-stat-label">Total Messages</div>
            </div>
            <div className="analytics-stat-item">
              <div className="analytics-stat-value">{stats.emojiReactions}</div>
              <div className="analytics-stat-label">Emoji Reactions</div>
            </div>
            <div className="analytics-stat-item">
              <div className="analytics-stat-value">{stats.directAccusations}</div>
              <div className="analytics-stat-label">Direct Accusations</div>
            </div>
            <div className="analytics-stat-item">
              <div className="analytics-stat-value">{stats.correctAIIdentifications}</div>
              <div className="analytics-stat-label">Correct AI IDs</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

export function PostGameAnalysis({ onBackToResults, onPlayAgain }: PostGameAnalysisProps) {
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
    <div className="analysis-container">
      <header className="analysis-header">
        <div className="company-logo">LOEBIAN</div>
        <div className="header-controls">
          <button 
            className="header-btn" 
            title="Toggle Theme"
            onClick={() => {
              const currentTheme = document.documentElement.getAttribute('data-theme');
              const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
              document.documentElement.setAttribute('data-theme', newTheme);
            }}
          >
            🌙
          </button>
        </div>
      </header>
      
      <div className="analysis-title-section">
        <h1 className="analysis-main-title">📊 POST-GAME ANALYSIS</h1>
        <p className="analysis-subtitle">Comprehensive review of operational events during the SEV-1 incident</p>
      </div>

      <div className="analysis-content">
        <div className="hud-tabs">
          <button 
            className={`hud-tab ${activeTab === 'summary' ? 'active' : ''}`}
            onClick={() => setActiveTab('summary')}
          >
            Summary
          </button>
          <button 
            className={`hud-tab ${activeTab === 'timeline' ? 'active' : ''}`}
            onClick={() => setActiveTab('timeline')}
          >
            Event Timeline
          </button>
          <button 
            className={`hud-tab ${activeTab === 'analytics' ? 'active' : ''}`}
            onClick={() => setActiveTab('analytics')}
          >
            Personnel Analytics
          </button>
          <button 
            className={`hud-tab ${activeTab === 'highlights' ? 'active' : ''}`}
            onClick={() => setActiveTab('highlights')}
          >
            Comms Highlights
          </button>
        </div>
        <div className="hud-content">
          {renderTabContent()}
        </div>
      </div>
      
      <div className="bottom-actions">
        <button className="btn-secondary" onClick={onBackToResults}>
          ← Back to Results
        </button>
        <button className="btn-primary" onClick={onPlayAgain}>
          🔄 Play Again
        </button>
      </div>
    </div>
  );
}