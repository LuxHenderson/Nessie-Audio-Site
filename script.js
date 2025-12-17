// script.js
// Site behaviors: smooth scroll, mobile menu, dark mode, lightbox, contact form,
// and local <video> initialization. This file intentionally avoids UI changes and
// focuses on progressive enhancement of existing markup.

(function(){
  'use strict';

  // Helper: select single or multiple elements
  const $ = (sel, all=false) => all ? Array.from(document.querySelectorAll(sel)) : document.querySelector(sel);

  // ---- Smooth scroll for internal links ----
  // Uses native scroll behavior; anchors with href starting with '#' will smoothly scroll.
  document.addEventListener('click', (e) => {
    const a = e.target.closest('a');
    if(!a) return;
    const href = a.getAttribute('href');
    if(href && href.startsWith('#')){
      const id = href.slice(1);
      const target = document.getElementById(id);
      if(target){
        e.preventDefault();
        target.scrollIntoView({behavior:'smooth', block:'start'});
        // Update history without jumping
        if(history.pushState) history.pushState(null, '', '#'+id);
      }
    }
  });

  // ---- Mobile menu toggle ----
  const menuToggle = $('.menu-toggle');
  const mobileMenu = $('#mobile-menu');
  if(menuToggle && mobileMenu){
    menuToggle.addEventListener('click', ()=>{
      const expanded = menuToggle.getAttribute('aria-expanded') === 'true';
      menuToggle.setAttribute('aria-expanded', String(!expanded));
      if(mobileMenu.hasAttribute('hidden')){
        mobileMenu.removeAttribute('hidden');
        menuToggle.textContent = 'Close';
      } else{
        mobileMenu.setAttribute('hidden','');
        menuToggle.textContent = 'Menu';
      }
    });
  }

  // Close mobile menu when clicking a link inside it
  if(mobileMenu){
    mobileMenu.addEventListener('click', (e)=>{
      const a = e.target.closest('a');
      if(a){
        mobileMenu.setAttribute('hidden','');
        menuToggle.setAttribute('aria-expanded','false');
        menuToggle.textContent = 'Menu';
      }
    });
  }

  // ---- Theme / Dark mode toggle ----
  const themeToggle = $('#theme-toggle');
  const root = document.documentElement;
  const savedTheme = localStorage.getItem('naevermore-theme');

  // Apply saved theme (if any) on load
  if(savedTheme) root.setAttribute('data-theme', savedTheme);

  // Update toggle label on load
  if(themeToggle){
    const isDark = (root.getAttribute('data-theme') === 'dark');
    themeToggle.textContent = isDark ? 'Light Mode' : 'Dark Mode';
    themeToggle.setAttribute('aria-pressed', String(isDark));
    themeToggle.addEventListener('click', ()=>{
      const current = root.getAttribute('data-theme');
      const next = (current === 'dark') ? '' : 'dark';
      if(next) root.setAttribute('data-theme', next); else root.removeAttribute('data-theme');
      localStorage.setItem('naevermore-theme', next);
      const isNowDark = (next === 'dark');
      themeToggle.textContent = isNowDark ? 'Light Mode' : 'Dark Mode';
      themeToggle.setAttribute('aria-pressed', String(isNowDark));
    });
  }

  // ---- Simple gallery lightbox ----
  // Opens a basic overlay with the full-size image. Closes on click/ESC.
  const thumbs = $('.thumb', true);
  const lightbox = $('#lightbox');
  const lightboxImg = $('#lightbox-img');
  const lbClose = $('#lightbox-close');

  function openLightbox(src, alt=''){
    if(!lightbox) return;
    lightboxImg.src = src;
    lightboxImg.alt = alt;
    lightbox.removeAttribute('hidden');
    lightbox.setAttribute('aria-hidden','false');
  }
  function closeLightbox(){
    if(!lightbox) return;
    lightbox.setAttribute('hidden','');
    lightbox.setAttribute('aria-hidden','true');
    lightboxImg.src = '';
  }

  thumbs.forEach(btn=>{
    btn.addEventListener('click', ()=>{
      const src = btn.dataset.full;
      const alt = btn.querySelector('img')?.alt || '';
      openLightbox(src, alt);
    });
  });

  if(lbClose) lbClose.addEventListener('click', closeLightbox);
  // close on backdrop click
  if(lightbox) lightbox.addEventListener('click', (e)=>{ if(e.target === lightbox) closeLightbox(); });
  // close on ESC
  document.addEventListener('keydown', (e)=>{ if(e.key === 'Escape') closeLightbox(); });

  // ---- Contact form: simple client-side validation/demo ----
  const form = $('#contact-form');
  const formMsg = $('#form-msg');
  if(form){
    form.addEventListener('submit', (e)=>{
      e.preventDefault();
      const data = new FormData(form);
      const name = data.get('name')?.toString().trim();
      const email = data.get('email')?.toString().trim();
      const message = data.get('message')?.toString().trim();

      // Basic checks; server-side required for real forms
      if(!name || !email || !message){
        formMsg.textContent = 'Please complete all fields.';
        formMsg.style.color = 'var(--accent)';
        return;
      }

      // Here you'd normally POST to your server or an email endpoint.
      // For this scaffold, we just show a success message and reset.
      formMsg.textContent = 'Thanks — your message has been queued (demo only).';
      formMsg.style.color = 'var(--muted)';
      form.reset();
    });
  }

  // ---- Booking form: Let Formspree handle submission directly (removed custom JS handler) ----
  // Form now submits naturally to Formspree without JavaScript intervention

  // ---- Small niceties ----
  // Keep copyright year up to date
  const year = $('#year'); if(year) year.textContent = new Date().getFullYear();

  // Scroll hint arrow: hide when user starts scrolling
  const scrollArrow = document.querySelector('.scroll-down-arrow');
  if(scrollArrow){
    function toggleArrow(){
      const y = window.scrollY || document.documentElement.scrollTop || 0;
      if(y > 20){ scrollArrow.classList.add('is-hidden'); }
      else{ scrollArrow.classList.remove('is-hidden'); }
    }
    // Initial check and on scroll/resize
    toggleArrow();
    window.addEventListener('scroll', toggleArrow, { passive: true });
    window.addEventListener('resize', toggleArrow);
  }

  // Enhanced search with dropdown results and cross-page navigation
  const search = $('#site-search');
  if(search){
    // Site-wide content map for search
    const siteContent = [
      { page: 'Nævermore.html', section: 'Home', keywords: 'home landing main page nævermore band' },
      { page: 'Nævermore.html', section: '- Upcoming Shows', keywords: 'upcoming shows concerts tour dates events gigs performances chester street canopy club memphis main' },
      { page: 'Nævermore.html', section: '- News', keywords: 'news updates announcements latest ashes echoes ep album release tour dates spring 2026' },
      { page: 'Nævermore.html', section: '- Follow Us', keywords: 'follow social media instagram facebook youtube tiktok' },
      { page: 'Nævermore.html', section: '- Booking', keywords: 'booking book us hire contact event venue date message inquiry request' },
      { page: 'about.html', section: 'About', keywords: 'about band members biography history story bio' },
      { page: 'music.html', section: 'Music', keywords: 'music songs tracks albums eps discography listen audio video ortus lux' },
      { page: 'merch.html', section: 'Merch', keywords: 'merch merchandise store shop buy clothing shirts apparel' },
      { page: 'tour.html', section: 'Tour', keywords: 'tour dates concerts shows schedule calendar tickets' },
      { page: 'gallery.html', section: 'Gallery', keywords: 'gallery photos pictures images media press' },
      { page: 'contact.html', section: 'Contact', keywords: 'contact email message reach us get in touch' }
    ];

    // Create dropdown container
    let dropdown = document.querySelector('.search-dropdown');
    if(!dropdown){
      dropdown = document.createElement('div');
      dropdown.className = 'search-dropdown';
      dropdown.setAttribute('role', 'listbox');
      search.parentElement.style.position = 'relative';
      search.parentElement.appendChild(dropdown);
    }

    search.addEventListener('input', ()=>{
      const q = search.value.trim().toLowerCase();
      
      // Clear dropdown if search is empty
      if(q === ''){ 
        dropdown.innerHTML = '';
        dropdown.style.display = 'none';
        return; 
      }
      
      // Search current page content
      const currentPageResults = [];
      const searchableElements = document.querySelectorAll('h1, h2, h3, h4, p, li, a, label');
      
      searchableElements.forEach(el=>{
        const text = el.innerText || el.textContent;
        if(text && text.toLowerCase().includes(q)){
          let section = el.closest('section, aside, .show-promo-row');
          let sectionTitle = null;
          if(section){
            const heading = section.querySelector('h1, h2, h3');
            if(heading) sectionTitle = heading.innerText.trim();
          }
          
          // Only add results that have a proper section title (skip generic "Page Content")
          if(sectionTitle){
            currentPageResults.push({
              section: sectionTitle,
              element: el,
              isCurrentPage: true
            });
          }
        }
      });
      
      // Search site-wide content map
      const siteResults = siteContent.filter(item=>{
        return item.section.toLowerCase().includes(q) || item.keywords.toLowerCase().includes(q);
      });
      
      // Combine and deduplicate results
      const allResults = [];
      const seen = new Set();
      
      // Add current page results first
      currentPageResults.forEach(r=>{
        const key = r.section;
        if(!seen.has(key)){
          seen.add(key);
          allResults.push(r);
        }
      });
      
      // Add site-wide results
      siteResults.forEach(r=>{
        const key = r.section + r.page;
        if(!seen.has(key)){
          seen.add(key);
          allResults.push(r);
        }
      });
      
      // Display results in dropdown - only show section titles (gold text)
      if(allResults.length > 0){
        dropdown.innerHTML = allResults.slice(0, 8).map((result, i)=>{
          return `<div class="search-result-item" data-index="${i}">
            <div class="result-section">${result.section}</div>
          </div>`;
        }).join('');
        dropdown.style.display = 'block';
        
        // Add click handlers to scroll to results or navigate to page
        dropdown.querySelectorAll('.search-result-item').forEach((item, i)=>{
          item.addEventListener('click', ()=>{
            const result = allResults[i];
            dropdown.style.display = 'none';
            search.value = '';
            
            if(result.isCurrentPage){
              // Scroll to element on current page
              result.element.scrollIntoView({ behavior: 'smooth', block: 'center' });
              result.element.style.backgroundColor = 'rgba(212, 175, 55, 0.2)';
              setTimeout(()=> result.element.style.backgroundColor = '', 2000);
            } else {
              // Navigate to different page
              window.location.href = result.page;
            }
          });
        });
      } else {
        dropdown.innerHTML = '<div class="search-no-results">No results found</div>';
        dropdown.style.display = 'block';
      }
    });
    
    // Close dropdown when clicking outside
    document.addEventListener('click', (e)=>{
      if(!search.contains(e.target) && !dropdown.contains(e.target)){
        dropdown.style.display = 'none';
      }
    });
    
    // Handle Enter key to navigate to first result
    search.addEventListener('keydown', (e)=>{
      if(e.key === 'Enter'){
        e.preventDefault();
        const firstResult = dropdown.querySelector('.search-result-item');
        if(firstResult){
          firstResult.click(); // Trigger click on first result
        }
      }
      if(e.key === 'Escape'){
        dropdown.style.display = 'none';
        search.value = '';
      }
    });
  }

  // ---- Local <video> loader ----
  // Finds elements with `data-video` and populates the <video> element. If no
  // `data-video` is present the section is hidden. Optional `data-poster`
  // provides a poster image. Keeps UI identical; only manages sources and fallback.
  Array.from(document.querySelectorAll('.video-wrap')).forEach(el=>{
    const file = (el.getAttribute('data-video') || '').trim();
    const poster = (el.getAttribute('data-poster') || '').trim();
    const section = el.closest('.video-section');
    if(!file){ if(section) section.style.display = 'none'; return; }

    let video = el.querySelector('video');
    if(!video){
      video = document.createElement('video');
      video.className = 'site-video';
      video.setAttribute('controls', '');
      video.setAttribute('playsinline', '');
      video.preload = 'metadata';
      el.insertBefore(video, el.firstChild);
    }

    // Populate source element (replace if present)
    let source = video.querySelector('source');
    if(source){ source.src = file; }
    else{
      source = document.createElement('source');
      source.src = file;
      // attempt to set type by extension (helps some browsers)
      const ext = file.split('.').pop().toLowerCase();
      if(ext === 'mp4') source.type = 'video/mp4';
      if(ext === 'webm') source.type = 'video/webm';
      video.appendChild(source);
    }

    if(poster) video.poster = poster;
    // show video once it can play, otherwise reveal fallback
    const fallback = el.querySelector('.video-fallback');
    const fallbackLink = el.querySelector('#video-fallback-link');
    if(fallbackLink){
      fallbackLink.href = file;
      fallbackLink.textContent = 'Open video file in new tab';
      fallbackLink.setAttribute('aria-label', 'Open video file');
    }

    function showFallback(){
      if(!fallback) return;
      // If a poster exists, show it as a linked thumbnail
      if(!el.querySelector('.video-thumb') && video.poster){
        const a = document.createElement('a');
        a.className = 'video-thumb';
        a.href = file;
        a.target = '_blank';
        a.rel = 'noopener noreferrer';
        const img = document.createElement('img');
        img.src = video.poster;
        img.alt = 'Open video in new tab';
        a.appendChild(img);
        fallback.insertBefore(a, fallback.firstChild);
      }
      fallback.hidden = false;
      video.hidden = true;
    }

    // When the video can play, show it and hide fallback
    video.addEventListener('canplay', ()=>{ if(fallback) fallback.hidden = true; video.hidden = false; });
    // If the browser fails to play, show fallback
    video.addEventListener('error', ()=>{ showFallback(); });

    // Attempt to load the media to trigger canplay/error
    try{ video.load(); }
    catch(e){ showFallback(); }
  });

})();
