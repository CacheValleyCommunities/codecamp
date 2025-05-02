// CodeCamp: Bridgerland - JavaScript

document.addEventListener('DOMContentLoaded', function() {
  // Mobile menu functionality
  const menuBtn = document.getElementById('menuBtn');
  const mobileMenu = document.getElementById('mobileMenu');
  
  if (menuBtn && mobileMenu) {
    menuBtn.addEventListener('click', function() {
      menuBtn.classList.toggle('active');
      mobileMenu.classList.toggle('active');
      document.body.classList.toggle('menu-open');
    });
    
    // Close mobile menu when clicking on a link
    const mobileLinks = mobileMenu.getElementsByTagName('a');
    for (let i = 0; i < mobileLinks.length; i++) {
      mobileLinks[i].addEventListener('click', function() {
        menuBtn.classList.remove('active');
        mobileMenu.classList.remove('active');
        document.body.classList.remove('menu-open');
      });
    }
  }
  
  // Add pixelated hover effect to buttons
  const retroButtons = document.querySelectorAll('.retro-btn');
  retroButtons.forEach(button => {
    button.addEventListener('mouseover', function() {
      // Add slight random position jitter for authentic retro feel
      const jitterX = Math.floor(Math.random() * 3) - 1;
      const jitterY = Math.floor(Math.random() * 3) - 1;
      this.style.transform = `translate(${jitterX}px, ${jitterY}px)`;
      
      // Play 8-bit hover sound if enabled
      playRetroSound('hover');
    });
    
    button.addEventListener('mouseout', function() {
      this.style.transform = 'translate(0, 0)';
    });
    
    button.addEventListener('click', function() {
      // Play 8-bit click sound if enabled
      playRetroSound('click');
    });
  });
  
  // Add text glitch effect on scroll
  const glitchElements = document.querySelectorAll('.glitch-text');
  let scrollTimeout;
  
  window.addEventListener('scroll', function() {
    glitchElements.forEach(el => {
      el.classList.add('active-glitch');
    });
    
    clearTimeout(scrollTimeout);
    scrollTimeout = setTimeout(function() {
      glitchElements.forEach(el => {
        el.classList.remove('active-glitch');
      });
    }, 300);
  });
  
  // Basic 8-bit sound effects (optional)
  function playRetroSound(type) {
    // Check if sound is enabled (could be a user preference)
    const soundEnabled = localStorage.getItem('retroSoundEnabled');
    if (soundEnabled !== 'true') return;
    
    const audioContext = new (window.AudioContext || window.webkitAudioContext)();
    const oscillator = audioContext.createOscillator();
    const gainNode = audioContext.createGain();
    
    oscillator.connect(gainNode);
    gainNode.connect(audioContext.destination);
    
    // Different sound for different actions
    if (type === 'hover') {
      oscillator.type = 'square';
      oscillator.frequency.value = 440;
      gainNode.gain.value = 0.1;
      oscillator.start();
      oscillator.stop(audioContext.currentTime + 0.1);
    } else if (type === 'click') {
      oscillator.type = 'square';
      oscillator.frequency.value = 660;
      gainNode.gain.value = 0.2;
      oscillator.start();
      oscillator.stop(audioContext.currentTime + 0.15);
    }
  }
  
  // Add typewriter effect for section titles
  function typewriterEffect() {
    const titles = document.querySelectorAll('.section-title');
    
    titles.forEach(title => {
      // Skip if already processed
      if (title.classList.contains('typed')) return;
      
      const observer = new IntersectionObserver(entries => {
        entries.forEach(entry => {
          if (entry.isIntersecting) {
            const text = title.textContent;
            title.textContent = '';
            title.classList.add('typed');
            
            let i = 0;
            const interval = setInterval(() => {
              if (i < text.length) {
                title.textContent += text.charAt(i);
                i++;
              } else {
                clearInterval(interval);
              }
            }, 100);
            
            observer.unobserve(title);
          }
        });
      }, { threshold: 0.5 });
      
      observer.observe(title);
    });
  }
  
  // Initialize typewriter effect with a slight delay
  setTimeout(typewriterEffect, 500);
  
  // Re-trigger typewriter effect on scroll
  window.addEventListener('scroll', function() {
    typewriterEffect();
  });
});