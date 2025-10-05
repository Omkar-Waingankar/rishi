import React, { useState, useEffect, useRef } from 'react';
import { Folder } from '@mui/icons-material';

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
              <Folder className="folder-icon" sx={{ fontSize: 12 }} />
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
