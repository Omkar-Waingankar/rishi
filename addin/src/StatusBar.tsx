import React, { useState, useEffect, useRef } from 'react';

interface StatusBarProps {
  connectionStatus: 'connecting' | 'connected' | 'failed';
  workingDirectory: string | null;
  triggerErrorRef?: React.MutableRefObject<(() => void) | null>;
}

// Truncate path to show only the last part with ellipsis
const truncatePath = (path: string, maxLength: number = 50): string => {
  if (path.length <= maxLength) {
    return path;
  }

  // Get the last directory/file name
  const parts = path.split('/');
  const lastPart = parts[parts.length - 1];

  // Show ".../" + last part
  const truncated = `…/${lastPart}`;

  // If still too long, truncate the last part too
  if (truncated.length > maxLength) {
    return `…/${lastPart.substring(0, maxLength - 5)}…`;
  }

  return truncated;
};

const StatusBar: React.FC<StatusBarProps> = ({ connectionStatus, workingDirectory, triggerErrorRef }) => {
  const [showTooltip, setShowTooltip] = useState<boolean>(false);
  const [isExpanded, setIsExpanded] = useState<boolean>(false);
  const [shouldShake, setShouldShake] = useState<boolean>(false);

  const hasError = connectionStatus === 'connected' && !workingDirectory;

  const triggerErrorAnimation = () => {
    // Trigger shake animation
    setShouldShake(true);
    setTimeout(() => setShouldShake(false), 500);

    // Expand to show full instructions (stays expanded until WD is set)
    setIsExpanded(true);
  };

  // Expose trigger method via ref
  useEffect(() => {
    if (triggerErrorRef) {
      triggerErrorRef.current = triggerErrorAnimation;
    }
    return () => {
      if (triggerErrorRef) {
        triggerErrorRef.current = null;
      }
    };
  }, [triggerErrorRef]);

  // Auto-collapse when working directory becomes valid
  useEffect(() => {
    if (workingDirectory && isExpanded) {
      setIsExpanded(false);
    }
  }, [workingDirectory, isExpanded]);

  return (
    <div className={`status-bar ${hasError ? 'error' : ''} ${isExpanded ? 'expanded' : ''} ${shouldShake ? 'shake' : ''}`}>
      {hasError ? (
        // Error state: no working directory
        <div className="status-bar-error-content">
          <div className="status-bar-left">
            <span className="status-dot failed"></span>
            <span className="status-text">⚠️ No working directory set</span>
          </div>
          {isExpanded && (
            <div className="status-bar-expanded-text">
              Open an .Rproj file or run <code>setwd("/path")</code> in the R console
            </div>
          )}
        </div>
      ) : (
        // Normal state
        <>
          <div className="status-bar-left">
            <span className={`status-dot ${connectionStatus}`}></span>
            <span className="status-text">
              {connectionStatus === 'connecting' && 'Connecting to server...'}
              {connectionStatus === 'connected' && 'Connected'}
              {connectionStatus === 'failed' && 'Failed to connect'}
            </span>
          </div>

          {connectionStatus === 'connected' && workingDirectory && (
            <div className="status-bar-right">
              <span className="status-separator">|</span>
              <svg className="folder-icon" width="12" height="12" viewBox="0 0 16 16" fill="currentColor">
                <path d="M1.75 1A1.75 1.75 0 000 2.75v10.5C0 14.216.784 15 1.75 15h12.5A1.75 1.75 0 0016 13.25v-8.5A1.75 1.75 0 0014.25 3H7.5a.25.25 0 01-.2-.1l-.9-1.2C6.07 1.26 5.55 1 5 1H1.75z"/>
              </svg>
              <span
                className="working-directory-text"
                onMouseEnter={() => setShowTooltip(true)}
                onMouseLeave={() => setShowTooltip(false)}
                title={workingDirectory}
              >
                {truncatePath(workingDirectory)}
              </span>
              {showTooltip && (
                <div className="working-directory-tooltip">
                  Working directory: {workingDirectory}
                  <br />
                  Change by opening an .Rproj file or running setwd() in the R console
                </div>
              )}
            </div>
          )}
        </>
      )}
    </div>
  );
};

export default StatusBar;
