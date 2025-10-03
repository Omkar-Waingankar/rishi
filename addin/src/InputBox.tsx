import React, { useState, useRef } from 'react';
import { InputBoxProps, ImageMimeType, MessageContent } from './types';
import ModelDropdown from './ModelDropdown';

interface DropdownOption {
  value: string;
  label: string;
}

interface ImageAttachment {
  id: string;
  file: File;
  previewUrl: string;
  mediaType: ImageMimeType;
  bytes: number;
}

const InputBox: React.FC<InputBoxProps> = ({ onSendMessage, disabled, isStreaming, onStopStreaming, safeRoot, triggerStatusBarError }) => {
  const [message, setMessage] = useState<string>('');
  const [selectedModel, setSelectedModel] = useState<string>('claude-4-sonnet');
  const [images, setImages] = useState<ImageAttachment[]>([]);
  const [imageError, setImageError] = useState<string | null>(null);
  const [hoveredImageId, setHoveredImageId] = useState<string | null>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  // Constants for validation
  const MAX_IMAGE_SIZE = 5 * 1024 * 1024; // 5MB
  const MAX_TOTAL_PAYLOAD = 20 * 1024 * 1024; // 20MB
  const MAX_IMAGES = 3;
  const SUPPORTED_TYPES: ImageMimeType[] = ['image/jpeg', 'image/png', 'image/webp', 'image/gif'];

  // Image validation helpers
  const validateImageType = (file: File): boolean => {
    return SUPPORTED_TYPES.includes(file.type as ImageMimeType);
  };

  const validateImageSize = (file: File): boolean => {
    return file.size <= MAX_IMAGE_SIZE;
  };

  const getTotalPayloadSize = (newImages: ImageAttachment[]): number => {
    return newImages.reduce((total, img) => total + img.bytes, 0);
  };

  // Convert file to base64
  const fileToBase64 = (file: File): Promise<string> => {
    return new Promise((resolve, reject) => {
      const reader = new FileReader();
      reader.onload = () => {
        const result = reader.result as string;
        // Strip the data URL prefix (e.g., "data:image/jpeg;base64,")
        const base64 = result.split(',')[1];
        resolve(base64);
      };
      reader.onerror = reject;
      reader.readAsDataURL(file);
    });
  };

  // Handle image addition
  const addImages = async (files: File[]): Promise<void> => {
    setImageError(null);
    
    // Calculate how many images we can actually add
    const availableSlots = MAX_IMAGES - images.length;
    const filesToProcess = files.slice(0, availableSlots);
    
    // If we're dropping some files due to limit, don't show error - just silently ignore extras
    
    const validFiles: File[] = [];
    for (const file of filesToProcess) {
      if (!validateImageType(file)) {
        setImageError(`Unsupported file type: ${file.type}. Only JPEG, PNG, WebP, and GIF are supported.`);
        return;
      }
      if (!validateImageSize(file)) {
        setImageError(`File too large: ${file.name}. Maximum size is 5MB per image.`);
        return;
      }
      validFiles.push(file);
    }

    try {
      const newAttachments: ImageAttachment[] = [];
      for (const file of validFiles) {
        const previewUrl = URL.createObjectURL(file);
        
        newAttachments.push({
          id: `${Date.now()}-${Math.random()}`,
          file,
          previewUrl,
          mediaType: file.type as ImageMimeType,
          bytes: file.size
        });
      }

      const updatedImages = [...images, ...newAttachments];
      if (getTotalPayloadSize(updatedImages) > MAX_TOTAL_PAYLOAD) {
        setImageError(`Total payload too large. Maximum ${MAX_TOTAL_PAYLOAD / (1024 * 1024)}MB allowed.`);
        // Clean up preview URLs
        newAttachments.forEach(img => URL.revokeObjectURL(img.previewUrl));
        return;
      }

      setImages(updatedImages);
    } catch (error) {
      console.error('Error processing images:', error);
      setImageError('Failed to process images. Please try again.');
    }
  };

  // Remove image
  const removeImage = (id: string): void => {
    setImages(prev => {
      const imageToRemove = prev.find(img => img.id === id);
      if (imageToRemove) {
        URL.revokeObjectURL(imageToRemove.previewUrl);
      }
      return prev.filter(img => img.id !== id);
    });
    setImageError(null);
  };

  // Handle file input change
  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>): void => {
    const files = Array.from(e.target.files || []);
    if (files.length > 0) {
      addImages(files);
    }
    // Reset input value to allow selecting the same file again
    e.target.value = '';
  };

  // Handle clipboard paste
  const handlePaste = async (e: React.ClipboardEvent<HTMLTextAreaElement>): Promise<void> => {
    const items = Array.from(e.clipboardData.items);
    const imageFiles: File[] = [];
    
    for (const item of items) {
      if (item.type.startsWith('image/')) {
        const file = item.getAsFile();
        if (file) {
          imageFiles.push(file);
        }
      }
    }
    
    if (imageFiles.length > 0) {
      e.preventDefault();
      await addImages(imageFiles);
    }
  };

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>): Promise<void> => {
    e.preventDefault();

    // If no safe root, trigger error animation instead of sending
    if (!safeRoot && (message.trim() || images.length > 0)) {
      if (triggerStatusBarError?.current) {
        triggerStatusBarError.current();
      }
      return;
    }

    if ((message.trim() || images.length > 0) && !disabled) {
      const content: MessageContent[] = [];
      
      // Add text content if present
      if (message.trim()) {
        content.push({
          type: 'text',
          content: message.trim()
        });
      }
      
      // Add image content
      for (const img of images) {
        try {
          const base64 = await fileToBase64(img.file);
          content.push({
            type: 'image',
            mediaType: img.mediaType,
            dataBase64: base64
          });
        } catch (error) {
          console.error('Error converting image to base64:', error);
          setImageError('Failed to process images. Please try again.');
          return;
        }
      }
      
      onSendMessage(content, selectedModel);
      setMessage('');
      setImages(prev => {
        // Clean up preview URLs
        prev.forEach(img => URL.revokeObjectURL(img.previewUrl));
        return [];
      });
      setImageError(null);
      if (textareaRef.current) {
        textareaRef.current.style.height = '24px';
      }
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>): void => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSubmit(e as unknown as React.FormEvent<HTMLFormElement>);
    }
  };

  const handleTextareaChange = (e: React.ChangeEvent<HTMLTextAreaElement>): void => {
    setMessage(e.target.value);
    
    const textarea = e.target;
    textarea.style.height = '24px';
    const scrollHeight = textarea.scrollHeight;
    const maxHeight = 120;
    textarea.style.height = Math.min(scrollHeight, maxHeight) + 'px';
  };

  // Clean up object URLs on unmount
  React.useEffect(() => {
    return () => {
      images.forEach(img => URL.revokeObjectURL(img.previewUrl));
    };
  }, []);

  const handleModelChange = (value: string): void => {
    setSelectedModel(value);
    localStorage.setItem('selectedModel', value);
  };

  const modelOptions: DropdownOption[] = [
    { value: 'claude-4-sonnet', label: 'Claude 4 Sonnet' },
    { value: 'claude-3.7-sonnet', label: 'Claude 3.7 Sonnet' }
  ];

  // Load saved model preference on component mount
  React.useEffect(() => {
    const savedModel = localStorage.getItem('selectedModel');
    if (savedModel) {
      setSelectedModel(savedModel);
    }
  }, []);

  return (
    <div className="input-box">
      <form onSubmit={handleSubmit}>
        <div className="input-container">
          {/* Image attachments strip */}
          {(images.length > 0 || imageError) && (
            <div className="attachment-strip">
              {imageError && (
                <div className="image-error">
                  <span className="image-error-text">{imageError}</span>
                  <button
                    type="button"
                    className="image-error-dismiss"
                    onClick={() => setImageError(null)}
                    aria-label="Dismiss error"
                  >
                    ×
                  </button>
                </div>
              )}
              {images.length > 0 && (
                <div className="image-previews">
                  {/* Hover preview */}
                  {hoveredImageId && (
                    <div className="image-hover-preview">
                      {(() => {
                        const hoveredImage = images.find(img => img.id === hoveredImageId);
                        return hoveredImage ? (
                          <img
                            src={hoveredImage.previewUrl}
                            alt={hoveredImage.file.name}
                            className="hover-preview-image"
                          />
                        ) : null;
                      })()}
                    </div>
                  )}
                  
                  {images.map(img => (
                    <div 
                      key={img.id} 
                      className="image-preview"
                      onMouseEnter={() => setHoveredImageId(img.id)}
                      onMouseLeave={() => setHoveredImageId(null)}
                      onClick={() => removeImage(img.id)}
                    >
                      <div className="image-thumbnail-container">
                        {hoveredImageId === img.id ? (
                          <div className="remove-x">×</div>
                        ) : (
                          <img
                            src={img.previewUrl}
                            alt={img.file.name}
                            className="image-thumbnail"
                          />
                        )}
                      </div>
                      <span className="image-label">Image</span>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}

          <div className="input-text-section">
            <textarea
              ref={textareaRef}
              value={message}
              onChange={handleTextareaChange}
              onKeyDown={handleKeyDown}
              onPaste={handlePaste}
              placeholder="Plan, build, analyze anything"
              disabled={disabled}
              rows={1}
            />
          </div>

          <div className="input-footer">
            <ModelDropdown
              value={selectedModel}
              onChange={handleModelChange}
              disabled={disabled}
              options={modelOptions}
            />

            <div className="input-actions">
              {/* Hidden file input */}
              <input
                ref={fileInputRef}
                type="file"
                accept="image/*"
                multiple
                onChange={handleFileSelect}
                style={{ display: 'none' }}
              />
              
              {/* Image upload button */}
              <button
                type="button"
                onClick={() => fileInputRef.current?.click()}
                disabled={disabled || images.length >= MAX_IMAGES}
                className="attachment-button"
                aria-label="Add image"
                title="Add image (or paste from clipboard)"
              >
                <svg className="attachment-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
                  <rect x="3" y="3" width="18" height="18" rx="2" ry="2"/>
                  <circle cx="9" cy="9" r="2"/>
                  <path d="m21 15-3.086-3.086a2 2 0 0 0-2.828 0L6 21"/>
                </svg>
              </button>

              {isStreaming ? (
                <button
                  type="button"
                  onClick={onStopStreaming}
                  className="send-button stop-button"
                  aria-label="Stop streaming"
                >
                  <svg className="send-icon" viewBox="0 0 24 24" fill="currentColor">
                    <rect x="6" y="6" width="12" height="12" rx="2"/>
                  </svg>
                </button>
              ) : (
                <button
                  type="submit"
                  disabled={disabled || (!message.trim() && images.length === 0)}
                  className="send-button"
                  aria-label="Send message"
                >
                  <svg className="send-icon" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M2,21L23,12L2,3V10L17,12L2,14V21Z"/>
                  </svg>
                </button>
              )}
            </div>
          </div>
        </div>
      </form>
    </div>
  );
};

export default InputBox;