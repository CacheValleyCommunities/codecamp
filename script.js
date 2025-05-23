// CodeCamp: Bridgerland - JavaScript

document.addEventListener("DOMContentLoaded", function () {
  // Global variables for music control
  let bgMusic = document.getElementById("bgMusic");
  let musicEnabled = false;
  let isMuted = false;

  function enableMusic() {
    musicEnabled = true;
    localStorage.setItem("musicPreference", "true");

    // Hide modal
    document.getElementById("musicModal").classList.remove("active");

    // Show music controls
    document.getElementById("musicControls").classList.add("active");

    // Set initial volume
    bgMusic.volume = 0.5;

    // Try to play music
    playBackgroundMusic();

    // Enable retro sound effects
    localStorage.setItem("retroSoundEnabled", "true");

    // Add hover sound effects to buttons
    addSoundEffects();
  }

  function skipMusic() {
    musicEnabled = false;
    localStorage.setItem("musicPreference", "false");
    localStorage.setItem("retroSoundEnabled", "false");

    // Hide modal
    document.getElementById("musicModal").classList.remove("active");
  }

  document.getElementById("enableMusic").addEventListener("click", enableMusic);
  document.getElementById("skipMusic").addEventListener("click", skipMusic);

  async function playBackgroundMusic() {
    try {
      await bgMusic.play();
    } catch (error) {
      console.log("Could not play background music:", error);
    }
  }

  function toggleMusic() {
    const playPauseBtn = document.getElementById("playPauseBtn");

    if (bgMusic.paused) {
      bgMusic.play();
      playPauseBtn.textContent = "⏸️";
      playPauseBtn.title = "Pause Music";
    } else {
      bgMusic.pause();
      playPauseBtn.textContent = "▶️";
      playPauseBtn.title = "Play Music";
    }
  }

  function toggleMute() {
    isMuted = !isMuted;
    bgMusic.muted = isMuted;

    const muteBtn = event.target;
    muteBtn.textContent = isMuted ? "🔇" : "🔊";
    muteBtn.title = isMuted ? "Unmute" : "Mute";
  }

  function setVolume(value) {
    bgMusic.volume = value / 100;
  }

  function playRetroSound(type) {
    const audioContext = new (window.AudioContext ||
      window.webkitAudioContext)();
    const oscillator = audioContext.createOscillator();
    const gainNode = audioContext.createGain();

    oscillator.connect(gainNode);
    gainNode.connect(audioContext.destination);

    // Different sound for different actions
    if (type === "hover") {
      oscillator.type = "square";
      oscillator.frequency.value = 440;
      gainNode.gain.value = 0.1;
      oscillator.start();
      oscillator.stop(audioContext.currentTime + 0.1);
    } else if (type === "click") {
      oscillator.type = "square";
      oscillator.frequency.value = 660;
      gainNode.gain.value = 0.2;
      oscillator.start();
      oscillator.stop(audioContext.currentTime + 0.15);
    }
  }

  function addSoundEffects() {
    // Add hover sounds to all buttons and interactive elements
    const interactiveElements = document.querySelectorAll(
      ".retro-btn, .music-btn, a",
    );

    interactiveElements.forEach((element) => {
      element.addEventListener("mouseenter", () => playRetroSound("hover"));
      element.addEventListener("click", () => playRetroSound("click"));
    });
  }

  // Handle audio loading errors gracefully
  bgMusic.addEventListener("error", function (e) {
    console.log("Background music failed to load, using fallback");
    if (musicEnabled) {
      createFallbackMusic();
    }
  });

  // Reset play button when music ends (in case it's not looping)
  bgMusic.addEventListener("ended", function () {
    document.getElementById("playPauseBtn").textContent = "▶️";
    document.getElementById("playPauseBtn").title = "Play Music";
  });

  // Mobile menu functionality
  const menuBtn = document.getElementById("menuBtn");
  const mobileMenu = document.getElementById("mobileMenu");

  if (menuBtn && mobileMenu) {
    menuBtn.addEventListener("click", function () {
      menuBtn.classList.toggle("active");
      mobileMenu.classList.toggle("active");
      document.body.classList.toggle("menu-open");
    });

    // Close mobile menu when clicking on a link
    const mobileLinks = mobileMenu.getElementsByTagName("a");
    for (let i = 0; i < mobileLinks.length; i++) {
      mobileLinks[i].addEventListener("click", function () {
        menuBtn.classList.remove("active");
        mobileMenu.classList.remove("active");
        document.body.classList.remove("menu-open");
      });
    }
  }

  // Add pixelated hover effect to buttons
  const retroButtons = document.querySelectorAll(".retro-btn");
  retroButtons.forEach((button) => {
    button.addEventListener("mouseover", function () {
      // Add slight random position jitter for authentic retro feel
      const jitterX = Math.floor(Math.random() * 3) - 1;
      const jitterY = Math.floor(Math.random() * 3) - 1;
      this.style.transform = `translate(${jitterX}px, ${jitterY}px)`;

      // Play 8-bit hover sound if enabled
      playRetroSound("hover");
    });

    button.addEventListener("mouseout", function () {
      this.style.transform = "translate(0, 0)";
    });

    button.addEventListener("click", function () {
      // Play 8-bit click sound if enabled
      playRetroSound("click");
    });
  });

  // Add text glitch effect on scroll
  const glitchElements = document.querySelectorAll(".glitch-text");
  let scrollTimeout;

  window.addEventListener("scroll", function () {
    glitchElements.forEach((el) => {
      el.classList.add("active-glitch");
    });

    clearTimeout(scrollTimeout);
    scrollTimeout = setTimeout(function () {
      glitchElements.forEach((el) => {
        el.classList.remove("active-glitch");
      });
    }, 300);
  });
});
