// fogEffect.js
// Cinematic atmospheric fog background using Three.js
// Runs passively behind page content with smooth fade-in

(function() {
  'use strict';

  // Wait for DOM and Three.js to be ready
  if (typeof THREE === 'undefined') {
    console.warn('Three.js not loaded. Fog effect will not initialize.');
    return;
  }

  // Configuration
  const config = {
    fogColor: 0x0a0a0f,
    fogNear: 1,
    fogFar: 15,
    initialFogDensity: 0,
    targetFogDensity: 0.15,
    fadeInDuration: 2500, // ms
    lightIntensity: 0.8,
    particleCount: 800, // More particles for fog-like appearance
    animationSpeed: 0.0003
  };

  let scene, camera, renderer, particles, animationId;
  let fogDensity = config.initialFogDensity;
  let lightIntensity = 0;
  let startTime = Date.now();

  // Initialize Three.js scene
  function init() {
    const container = document.getElementById('fog-canvas-container');
    if (!container) {
      console.warn('Fog canvas container not found');
      return;
    }

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
      powerPreference: 'low-power' // Prefer low GPU usage
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

    // Start animation loop
    animate();
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

  // Animation loop
  function animate() {
    animationId = requestAnimationFrame(animate);

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
    window.removeEventListener('resize', onWindowResize);
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

  // Initialize when DOM is ready
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }

  // Expose destroy for cleanup if needed
  window.fogEffectDestroy = destroy;

})();
