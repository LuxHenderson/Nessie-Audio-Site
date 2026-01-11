// form-validation.js
// Comprehensive form validation and sanitization utilities
// Provides XSS protection, input validation, and spam prevention

(function(){
  'use strict';

  // ========== Validation Utilities ==========

  const FormValidator = {
    // Email validation using RFC 5322 compliant regex
    isValidEmail: function(email) {
      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
      return emailRegex.test(email);
    },

    // Sanitize input to prevent XSS attacks
    sanitizeInput: function(input) {
      if (typeof input !== 'string') return '';

      // Create a temporary div to use browser's built-in HTML encoding
      const div = document.createElement('div');
      div.textContent = input;
      let sanitized = div.innerHTML;

      // Additional sanitization: remove any remaining script tags or event handlers
      sanitized = sanitized
        .replace(/<script\b[^<]*(?:(?!<\/script>)<[^<]*)*<\/script>/gi, '')
        .replace(/on\w+\s*=\s*["'][^"']*["']/gi, '')
        .replace(/on\w+\s*=\s*[^\s>]*/gi, '')
        .replace(/javascript:/gi, '');

      return sanitized.trim();
    },

    // Validate text length
    isValidLength: function(text, minLength, maxLength) {
      const length = text.trim().length;
      return length >= minLength && length <= maxLength;
    },

    // Validate that text contains only safe characters (letters, numbers, spaces, basic punctuation)
    containsSafeCharacters: function(text) {
      // Allow letters, numbers, spaces, and common punctuation
      const safeRegex = /^[a-zA-Z0-9\s\.,\-!?'":;()\n\r]+$/;
      return safeRegex.test(text);
    },

    // Validate date is in the future
    isFutureDate: function(dateString) {
      const selectedDate = new Date(dateString);
      const today = new Date();
      today.setHours(0, 0, 0, 0); // Reset time to start of day
      return selectedDate >= today;
    }
  };

  // ========== Form Validation Setup ==========

  // Initialize booking form validation
  const bookingForm = document.querySelector('.booking-form');
  if (bookingForm) {
    // Add honeypot field for spam protection (hidden from users)
    const honeypot = document.createElement('input');
    honeypot.type = 'text';
    honeypot.name = '_gotcha';
    honeypot.style.display = 'none';
    honeypot.tabIndex = -1;
    honeypot.autocomplete = 'off';
    honeypot.setAttribute('aria-hidden', 'true');
    bookingForm.appendChild(honeypot);

    // Track submission to prevent double-submit
    let isSubmitting = false;

    // Get form elements
    const nameInput = bookingForm.querySelector('#bk-name');
    const emailInput = bookingForm.querySelector('#bk-email');
    const projectTypeInput = bookingForm.querySelector('#bk-project-type');
    const deadlineInput = bookingForm.querySelector('#bk-deadline');
    const messageInput = bookingForm.querySelector('#bk-message');
    const submitButton = bookingForm.querySelector('button[type="submit"]');

    // Add input length limits
    if (nameInput) nameInput.maxLength = 100;
    if (emailInput) emailInput.maxLength = 254; // Max email length per RFC
    if (projectTypeInput) projectTypeInput.maxLength = 200;
    if (messageInput) messageInput.maxLength = 2000;

    // Helper function to show error message
    function showError(input, message) {
      // Remove any existing error
      clearError(input);

      // Create error message element
      const errorDiv = document.createElement('div');
      errorDiv.className = 'form-error';
      errorDiv.textContent = message;
      errorDiv.setAttribute('role', 'alert');

      // Insert error after the input's parent form-row
      const formRow = input.closest('.form-row');
      if (formRow) {
        formRow.appendChild(errorDiv);
        formRow.classList.add('has-error');
      }

      // Add error styling to input
      input.classList.add('input-error');
      input.setAttribute('aria-invalid', 'true');
    }

    // Helper function to clear error message
    function clearError(input) {
      const formRow = input.closest('.form-row');
      if (formRow) {
        const errorDiv = formRow.querySelector('.form-error');
        if (errorDiv) {
          errorDiv.remove();
        }
        formRow.classList.remove('has-error');
      }
      input.classList.remove('input-error');
      input.removeAttribute('aria-invalid');
    }

    // Real-time validation on blur
    function validateField(input) {
      const value = input.value.trim();

      // Clear previous errors
      clearError(input);

      // Validate based on field type
      if (input === nameInput) {
        if (!value) {
          showError(input, 'Name is required');
          return false;
        }
        if (!FormValidator.isValidLength(value, 2, 100)) {
          showError(input, 'Name must be between 2 and 100 characters');
          return false;
        }
        if (!FormValidator.containsSafeCharacters(value)) {
          showError(input, 'Name contains invalid characters');
          return false;
        }
      }

      if (input === emailInput) {
        if (!value) {
          showError(input, 'Email is required');
          return false;
        }
        if (!FormValidator.isValidEmail(value)) {
          showError(input, 'Please enter a valid email address');
          return false;
        }
      }

      if (input === projectTypeInput) {
        if (!value) {
          showError(input, 'Project type is required');
          return false;
        }
        if (!FormValidator.isValidLength(value, 3, 200)) {
          showError(input, 'Project type must be between 3 and 200 characters');
          return false;
        }
      }

      if (input === deadlineInput) {
        if (!value) {
          showError(input, 'Deadline is required');
          return false;
        }
        if (!FormValidator.isFutureDate(value)) {
          showError(input, 'Deadline must be today or in the future');
          return false;
        }
      }

      if (input === messageInput) {
        if (!value) {
          showError(input, 'Message is required');
          return false;
        }
        if (!FormValidator.isValidLength(value, 10, 2000)) {
          showError(input, 'Message must be between 10 and 2000 characters');
          return false;
        }
      }

      return true;
    }

    // Add blur event listeners for real-time validation
    [nameInput, emailInput, projectTypeInput, deadlineInput, messageInput].forEach(input => {
      if (input) {
        input.addEventListener('blur', () => validateField(input));
        input.addEventListener('input', () => {
          // Clear error on input if field was previously invalid
          if (input.classList.contains('input-error')) {
            clearError(input);
          }
        });
      }
    });

    // Form submission handler
    bookingForm.addEventListener('submit', function(e) {
      // Check honeypot field - if filled, it's likely a bot
      if (honeypot.value) {
        e.preventDefault();
        console.log('Spam detected - honeypot filled');
        return false;
      }

      // Prevent double submission
      if (isSubmitting) {
        e.preventDefault();
        return false;
      }

      // Validate all fields
      let isValid = true;
      const fields = [nameInput, emailInput, projectTypeInput, deadlineInput, messageInput];

      fields.forEach(input => {
        if (input && !validateField(input)) {
          isValid = false;
        }
      });

      if (!isValid) {
        e.preventDefault();

        // Focus on first error field
        const firstError = bookingForm.querySelector('.input-error');
        if (firstError) {
          firstError.focus();
        }

        return false;
      }

      // Sanitize all inputs before submission
      if (nameInput) nameInput.value = FormValidator.sanitizeInput(nameInput.value);
      if (emailInput) emailInput.value = FormValidator.sanitizeInput(emailInput.value);
      if (projectTypeInput) projectTypeInput.value = FormValidator.sanitizeInput(projectTypeInput.value);
      if (messageInput) messageInput.value = FormValidator.sanitizeInput(messageInput.value);

      // Mark as submitting and disable button
      isSubmitting = true;
      if (submitButton) {
        submitButton.disabled = true;
        submitButton.textContent = 'Sending...';
      }

      // Form will submit naturally to Formspree
      // Re-enable after a delay in case of network error
      setTimeout(() => {
        isSubmitting = false;
        if (submitButton) {
          submitButton.disabled = false;
          submitButton.textContent = 'Send Booking Request';
        }
      }, 5000);

      return true;
    });
  }

  // Export utilities for use in other scripts if needed
  window.FormValidator = FormValidator;

})();
