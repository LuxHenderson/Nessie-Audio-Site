// fogEffect.js
// Cinematic atmospheric fog background using Three.js
// Runs passively behind page content with smooth fade-in

(function() {
  'use strict';

  // Wait for Three.js to load (may still be downloading from CDN on first visit)
  function waitForThree(callback, attempts) {
    if (typeof THREE !== 'undefined') {
      callback();
      return;
    }
    if (attempts <= 0) {
      console.warn('Three.js failed to load. Fog effect will not initialize.');
      return;
    }
    setTimeout(function() { waitForThree(callback, attempts - 1); }, 100);
  }

  // Configuration
  const config = {
    fogColor: 0x303040,
    fogNear: 1,
    fogFar: 15,
    initialFogDensity: 0.08, // Start slightly visible instead of 0
    targetFogDensity: 0.2,
    fadeInDuration: 1500, // Slightly faster fade-in
    lightIntensity: 1.2,
    particleCount: 800, // More particles for fog-like appearance
    animationSpeed: 0.0003
  };

  let scene, camera, renderer, particles, animationId;
  let fogDensity = config.initialFogDensity;
  let lightIntensity = 0;
  let startTime = Date.now();
  let isPageVisible = true;
  let lastTime = Date.now();
  let lastFrameTime = Date.now();
  let watchdogInterval = null;

  // Initialize Three.js scene
  function init() {
    const container = document.getElementById('fog-canvas-container');
    if (!container) {
      console.warn('Fog canvas container not found');
      return;
    }

    try {
      // Reset start time so fade-in begins from actual render start, not script parse time
      startTime = Date.now();

      // Scene setup
      scene = new THREE.Scene();
      scene.fog = new THREE.FogExp2(config.fogColor, fogDensity);

      // Camera setup - wide FOV for immersive feel
      camera = new THREE.PerspectiveCamera(
        75,
        window.innerWidth / window.innerHeight,
        0.1,
        1000
      );
      camera.position.z = 5;

      // Renderer setup
      renderer = new THREE.WebGLRenderer({
        alpha: true,
        antialias: false, // Disable for performance
        powerPreference: 'high-performance' // Changed from 'low-power' to prevent Chrome from releasing GPU resources
      });
      renderer.setSize(window.innerWidth, window.innerHeight);
      renderer.setPixelRatio(Math.min(window.devicePixelRatio, 1.5)); // Cap pixel ratio for performance
      container.appendChild(renderer.domElement);

      // Create atmospheric particles/fog elements
      createFogParticles();

      // Add subtle ambient light
      const ambientLight = new THREE.AmbientLight(0x404050, 0);
      scene.add(ambientLight);
      
      // Add directional light for depth
      const directionalLight = new THREE.DirectionalLight(0x8888aa, 0);
      directionalLight.position.set(0, 1, 0.5);
      scene.add(directionalLight);

      // Store lights for animation
      scene.userData.ambientLight = ambientLight;
      scene.userData.directionalLight = directionalLight;

      // Window resize handler
      window.addEventListener('resize', onWindowResize, false);

      // Visibility change detection - pause/resume animation for Chrome tab throttling
      document.addEventListener('visibilitychange', handleVisibilityChange, false);
      window.addEventListener('blur', handleWindowBlur, false);
      window.addEventListener('focus', handleWindowFocus, false);

      // Start watchdog to detect if Chrome completely stops RAF
      startWatchdog();

      // Start animation loop
      animate();
    } catch (error) {
      console.error('Error initializing fog effect:', error);
    }
  }

  // Create fog particles for volumetric effect
  function createFogParticles() {
    const geometry = new THREE.BufferGeometry();
    const positions = [];
    const velocities = [];
    const sizes = [];
    const opacities = [];

    for (let i = 0; i < config.particleCount; i++) {
      // Distribute particles in 3D space - wider, more natural spread
      positions.push(
        (Math.random() - 0.5) * 50,
        (Math.random() - 0.5) * 30,
        (Math.random() - 0.5) * 40 - 10
      );

      // Very slow random drift velocities for fog-like movement
      velocities.push(
        (Math.random() - 0.5) * 0.008,
        (Math.random() - 0.5) * 0.004,
        (Math.random() - 0.5) * 0.008
      );

      // Varied sizes for natural fog appearance
      sizes.push(Math.random() * 2 + 1.5);
      
      // Varied opacities
      opacities.push(Math.random() * 0.3 + 0.1);
    }

    geometry.setAttribute('position', new THREE.Float32BufferAttribute(positions, 3));
    geometry.setAttribute('size', new THREE.Float32BufferAttribute(sizes, 1));
    geometry.setAttribute('opacity', new THREE.Float32BufferAttribute(opacities, 1));
    geometry.userData.velocities = velocities;

    // Custom shader material for realistic fog appearance
    const material = new THREE.ShaderMaterial({
      uniforms: {
        time: { value: 0 },
        baseOpacity: { value: 0.0 }
      },
      vertexShader: `
        attribute float size;
        attribute float opacity;
        varying float vOpacity;
        
        void main() {
          vOpacity = opacity;
          vec4 mvPosition = modelViewMatrix * vec4(position, 1.0);
          gl_PointSize = size * (300.0 / -mvPosition.z);
          gl_Position = projectionMatrix * mvPosition;
        }
      `,
      fragmentShader: `
        uniform float baseOpacity;
        varying float vOpacity;
        
        void main() {
          // Soft circular gradient for each particle
          vec2 center = gl_PointCoord - vec2(0.5);
          float dist = length(center);
          float alpha = 1.0 - smoothstep(0.0, 0.5, dist);
          alpha = alpha * alpha; // Softer falloff
          
          // Misty blue-gray color
          vec3 color = vec3(0.4, 0.45, 0.55);
          
          gl_FragColor = vec4(color, alpha * vOpacity * baseOpacity);
        }
      `,
      transparent: true,
      depthWrite: false,
      blending: THREE.NormalBlending
    });

    particles = new THREE.Points(geometry, material);
    scene.add(particles);
  }

  // Smooth fade-in animation using easing
  function easeOutCubic(t) {
    return 1 - Math.pow(1 - t, 3);
  }

  // Watchdog timer to detect if RAF has been completely stopped by Chrome
  function startWatchdog() {
    // Check every 2 seconds if animation is still running
    watchdogInterval = setInterval(() => {
      const now = Date.now();
      const timeSinceLastFrame = now - lastFrameTime;

      // If more than 3 seconds since last frame AND page is visible, restart
      if (timeSinceLastFrame > 3000 && !document.hidden) {
        console.log('Fog animation watchdog: Restarting animation after Chrome throttle');
        if (animationId) {
          cancelAnimationFrame(animationId);
        }
        animationId = null;
        lastFrameTime = now;
        animate();
      }
    }, 2000);
  }

  // Handle visibility change (tab switching, minimizing)
  function handleVisibilityChange() {
    if (document.hidden) {
      // Tab is now hidden - just mark as not visible
      // DON'T cancel RAF - let it throttle naturally
      isPageVisible = false;
    } else {
      // Tab is now visible - ensure animation is running
      isPageVisible = true;
      lastTime = Date.now(); // Reset time tracking to prevent jumps

      // CRITICAL: Always restart animation when becoming visible
      // Chrome may have completely stopped RAF/setInterval when minimized
      const timeSinceLastFrame = Date.now() - lastFrameTime;

      if (!animationId || timeSinceLastFrame > 1000) {
        // Animation stopped or stalled - force restart
        if (animationId) {
          cancelAnimationFrame(animationId);
        }
        animationId = null;
        lastFrameTime = Date.now();
        console.log('Visibility restored: Restarting fog animation');
        animate();
      }
    }
  }

  // Handle window blur (user clicked away) - be less aggressive
  function handleWindowBlur() {
    // Just mark as not visible, don't stop the loop
    isPageVisible = false;
  }

  // Handle window focus (user came back) - ensure animation restarts
  function handleWindowFocus() {
    isPageVisible = true;
    lastTime = Date.now();

    // Critical: Check if animation has stalled
    const timeSinceLastFrame = Date.now() - lastFrameTime;

    if (!animationId || timeSinceLastFrame > 1000) {
      // Animation stopped or stalled - force restart
      if (animationId) {
        cancelAnimationFrame(animationId);
      }
      animationId = null;
      lastFrameTime = Date.now();
      console.log('Window focus restored: Restarting fog animation');
      animate();
    }
  }

  // Animation loop
  function animate() {
    animationId = requestAnimationFrame(animate);

    // Update last frame time for watchdog
    lastFrameTime = Date.now();

    const elapsed = Date.now() - startTime;
    const fadeProgress = Math.min(elapsed / config.fadeInDuration, 1);
    const easedProgress = easeOutCubic(fadeProgress);

    // Animate fog density fade-in
    if (fadeProgress < 1) {
      fogDensity = config.initialFogDensity + 
        (config.targetFogDensity - config.initialFogDensity) * easedProgress;
      scene.fog.density = fogDensity;

      // Animate light intensity
      lightIntensity = config.lightIntensity * easedProgress;
      scene.userData.ambientLight.intensity = lightIntensity * 0.5;
      scene.userData.directionalLight.intensity = lightIntensity * 0.3;
      
      // Fade in particle opacity
      if (particles && particles.material.uniforms) {
        particles.material.uniforms.baseOpacity.value = easedProgress;
      }
    }

    // Animate particles - slow drift
    if (particles) {
      const positions = particles.geometry.attributes.position.array;
      const velocities = particles.geometry.userData.velocities;

      for (let i = 0; i < positions.length; i += 3) {
        positions[i] += velocities[i];
        positions[i + 1] += velocities[i + 1];
        positions[i + 2] += velocities[i + 2];

        // Wrap particles around bounds for seamless effect
        if (Math.abs(positions[i]) > 25) positions[i] *= -0.95;
        if (Math.abs(positions[i + 1]) > 15) positions[i + 1] *= -0.95;
        if (Math.abs(positions[i + 2]) > 20) positions[i + 2] *= -0.95;
      }

      particles.geometry.attributes.position.needsUpdate = true;
      
      // Very subtle rotation for organic movement
      particles.rotation.y += config.animationSpeed * 0.5;
      
      // Update shader time uniform for any time-based effects
      if (particles.material.uniforms) {
        particles.material.uniforms.time.value = elapsed * 0.001;
      }
    }

    // Camera subtle movement for depth - very gentle
    camera.position.x = Math.sin(elapsed * 0.00008) * 0.3;
    camera.position.y = Math.cos(elapsed * 0.0001) * 0.2;

    renderer.render(scene, camera);
  }

  // Handle window resize
  function onWindowResize() {
    camera.aspect = window.innerWidth / window.innerHeight;
    camera.updateProjectionMatrix();
    renderer.setSize(window.innerWidth, window.innerHeight);
  }

  // Cleanup function
  function destroy() {
    if (animationId) {
      cancelAnimationFrame(animationId);
    }
    if (watchdogInterval) {
      clearInterval(watchdogInterval);
    }
    window.removeEventListener('resize', onWindowResize);
    document.removeEventListener('visibilitychange', handleVisibilityChange);
    window.removeEventListener('blur', handleWindowBlur);
    window.removeEventListener('focus', handleWindowFocus);
    if (renderer) {
      renderer.dispose();
      const container = document.getElementById('fog-canvas-container');
      if (container && renderer.domElement) {
        container.removeChild(renderer.domElement);
      }
    }
    if (particles) {
      particles.geometry.dispose();
      particles.material.dispose();
    }
  }

  // Initialize when DOM is ready, waiting for Three.js if needed
  function start() {
    // Guard against double-initialization
    if (animationId) return;
    waitForThree(init, 50); // Retry up to 50 times (5 seconds)
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', start);
  } else {
    start();
  }

  // Fallback: if defer script timing was disrupted (e.g. by a 301 redirect),
  // ensure fog initializes once the page is fully loaded
  window.addEventListener('load', function() {
    if (!animationId && document.getElementById('fog-canvas-container')) {
      start();
    }
  });

  // Handle bfcache restoration (browser restores page state without re-executing scripts)
  window.addEventListener('pageshow', function(event) {
    if (event.persisted && !animationId) {
      start();
    }
  });

  // Expose destroy for cleanup if needed
  window.fogEffectDestroy = destroy;

})();
