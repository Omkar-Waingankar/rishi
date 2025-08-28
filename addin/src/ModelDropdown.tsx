import React, { useState, useRef } from 'react';

interface DropdownOption {
  value: string;
  label: string;
}

interface ModelDropdownProps {
  value: string;
  onChange: (value: string) => void;
  disabled: boolean;
  options: DropdownOption[];
}

const ModelDropdown: React.FC<ModelDropdownProps> = ({ value, onChange, disabled, options }) => {
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  const selectedOption = options.find(option => option.value === value);

  const handleOptionClick = (optionValue: string) => {
    onChange(optionValue);
    setIsOpen(false);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Escape') {
      setIsOpen(false);
    }
  };

  // Close dropdown when clicking outside
  React.useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };

    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [isOpen]);

  return (
    <div className="custom-dropdown" ref={dropdownRef} onKeyDown={handleKeyDown}>
      <button
        type="button"
        className={`dropdown-trigger ${isOpen ? 'open' : ''}`}
        onClick={() => !disabled && setIsOpen(!isOpen)}
        disabled={disabled}
        aria-label="Select AI model"
        aria-expanded={isOpen}
        aria-haspopup="listbox"
      >
        <span className="dropdown-value">{selectedOption?.label}</span>
        <svg 
          className="dropdown-arrow" 
          width="12" 
          height="12" 
          viewBox="0 0 12 12" 
          fill="none"
        >
          <path 
            d="M3 7.5L6 4.5L9 7.5" 
            stroke="currentColor" 
            strokeWidth="1.5" 
            strokeLinecap="round" 
            strokeLinejoin="round"
          />
        </svg>
      </button>
      
      {isOpen && (
        <div className="dropdown-menu" role="listbox">
          {options.map((option) => (
            <button
              key={option.value}
              type="button"
              className={`dropdown-option ${option.value === value ? 'selected' : ''}`}
              onClick={() => handleOptionClick(option.value)}
              role="option"
              aria-selected={option.value === value}
            >
              {option.label}
            </button>
          ))}
        </div>
      )}
    </div>
  );
};

export default ModelDropdown;
