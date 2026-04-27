import { WelcomePane } from "@/components/terminal/welcome-pane";
import { ForumActivityPane } from "@/components/terminal/forum-activity-pane";
import { LanguagesPane } from "@/components/terminal/languages-pane";
import { StatsPane } from "@/components/terminal/stats-pane";

export function HomeContent() {
  return (
    <div className="terminal-home-grid">
      <div style={{ order: 1 }}>
        <WelcomePane />
      </div>
      <div style={{ order: 2 }}>
        <ForumActivityPane />
      </div>
      <div style={{ order: 3 }}>
        <LanguagesPane />
      </div>
      <div style={{ order: 4 }}>
        <StatsPane />
      </div>
    </div>
  );
}
