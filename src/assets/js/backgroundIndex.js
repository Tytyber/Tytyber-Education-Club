/* Немного JS — фон на canvas: плавающие точки + мягкие волны.
   Файл независим, не требует внешних библиотек.
*/

(() => {
  const canvas = document.getElementById('bgCanvas');
  const ctx = canvas.getContext('2d', { alpha: true });
  let w = 0, h = 0, DPR = window.devicePixelRatio || 1;

  function resize() {
    DPR = window.devicePixelRatio || 1;
    w = window.innerWidth;
    h = window.innerHeight;
    canvas.width = Math.floor(w * DPR);
    canvas.height = Math.floor(h * DPR);
    canvas.style.width = w + 'px';
    canvas.style.height = h + 'px';
    ctx.setTransform(DPR, 0, 0, DPR, 0, 0);
  }
  window.addEventListener('resize', resize);
  resize();

  // Простая система частиц
  const particles = [];
  const PARTICLE_COUNT = Math.round((w * h) / 50000); // масштабируемое кол-во
  const max = Math.max;

  function rand(min, max) { return Math.random() * (max - min) + min; }

  class Particle {
    constructor() {
      this.reset();
    }
    reset() {
      this.x = rand(0, w);
      this.y = rand(0, h);
      this.r = rand(0.8, 3.5); // радиус
      this.vx = rand(-0.2, 0.2);
      this.vy = rand(-0.1, -0.6);
      this.alpha = rand(0.05, 0.8);
      this.phase = rand(0, Math.PI * 2);
      this.speed = rand(0.2, 1.2);
    }
    update(t) {
      this.x += this.vx * this.speed;
      this.y += this.vy * this.speed - Math.sin((t + this.phase) * 0.001) * 0.2;
      this.phase += 0.005;
      if (this.y < -50 || this.x < -50 || this.x > w + 50) this.reset();
    }
    draw(ctx) {
      ctx.save();
      ctx.beginPath();
      ctx.fillStyle = `rgba(57,255,20, ${this.alpha})`;
      // glow
      ctx.shadowColor = 'rgba(57,255,20,0.55)';
      ctx.shadowBlur = 18;
      ctx.arc(this.x, this.y, this.r, 0, Math.PI * 2);
      ctx.fill();
      ctx.restore();
    }
  }

  // Инициализация частиц
  function initParticles(){
    particles.length = 0;
    const count = max(12, PARTICLE_COUNT);
    for (let i = 0; i < count; i++) particles.push(new Particle());
  }
  initParticles();

  // Рисуем плавные волны (несколько синусоид)
  function drawWaves(t) {
    // overlay gradient
    const g = ctx.createLinearGradient(0, 0, w, h);
    g.addColorStop(0, 'rgba(0, 10, 0, 0.32)');
    g.addColorStop(1, 'rgba(0, 40, 15, 0.52)');
    ctx.fillStyle = g;
    ctx.fillRect(0, 0, w, h);

    // волны
    for (let k = 0; k < 3; k++) {
      ctx.beginPath();
      const amplitude = 20 + k * 18;
      const frequency = 0.0015 + k * 0.0008;
      ctx.moveTo(0, h);
      for (let x = 0; x <= w; x += 12) {
        const y = h * (0.6 - k * 0.07) + Math.sin((x + t * (0.25 + k * 0.15)) * frequency) * amplitude;
        ctx.lineTo(x, y);
      }
      ctx.lineTo(w, h);
      ctx.closePath();
      ctx.fillStyle = `rgba(0,255,127,${0.02 + k * 0.02})`;
      ctx.fill();
    }
  }

  // Подсветка в центре (радиальный градиент)
  function drawCenterGlow(t){
    const cx = w * 0.18;
    const cy = h * 0.25;
    const rg = ctx.createRadialGradient(cx, cy, 0, cx, cy, Math.max(w, h) * 0.8);
    rg.addColorStop(0, 'rgba(57,255,20,0.12)');
    rg.addColorStop(0.25, 'rgba(57,255,20,0.06)');
    rg.addColorStop(1, 'rgba(0,0,0,0)');
    ctx.fillStyle = rg;
    ctx.fillRect(0, 0, w, h);
  }

  let last = performance.now();

  function loop(now) {
    const t = now;
    const dt = now - last;
    last = now;

    // clear
    ctx.clearRect(0, 0, w, h);

    // background base
    // dark gradient base
    const bg = ctx.createLinearGradient(0, 0, 0, h);
    bg.addColorStop(0, '#021503');
    bg.addColorStop(1, '#041b06');
    ctx.fillStyle = bg;
    ctx.fillRect(0, 0, w, h);

    // center glow and waves
    drawCenterGlow(t);
    drawWaves(t);

    // update & draw particles
    for (let p of particles) {
      p.update(t);
      p.draw(ctx);
    }

    // мягкие линейные соединения между близкими частицами
    for (let i = 0; i < particles.length; i++) {
      for (let j = i + 1; j < particles.length; j++) {
        const a = particles[i], b = particles[j];
        const dx = a.x - b.x, dy = a.y - b.y;
        const d = Math.sqrt(dx * dx + dy * dy);
        if (d < 120) {
          ctx.beginPath();
          ctx.moveTo(a.x, a.y);
          ctx.lineTo(b.x, b.y);
          ctx.strokeStyle = `rgba(0,255,127,${0.06 * (1 - d / 120)})`;
          ctx.lineWidth = 1;
          ctx.stroke();
        }
      }
    }

    requestAnimationFrame(loop);
  }

  requestAnimationFrame(loop);

  // optional: интерактивность — небольшой сдвиг частиц по движению мыши
  let mouse = {x: w/2, y: h/2};
  window.addEventListener('mousemove', (e) => {
    mouse.x = e.clientX;
    mouse.y = e.clientY;
    // лёгкий отклик: смещаем ближе тех, кто рядом
    for (let p of particles) {
      const dx = p.x - mouse.x;
      const dy = p.y - mouse.y;
      const dist = Math.sqrt(dx*dx + dy*dy);
      if (dist < 120) {
        p.x += (dx / dist) * 2;
        p.y += (dy / dist) * 2;
      }
    }
  });

  // пересоздать частицы при изменении размера
  window.addEventListener('resize', () => {
    initParticles();
  });

})();
