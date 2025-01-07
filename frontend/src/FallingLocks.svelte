<!-- FallingLocks.svelte -->
<script>
    import { onMount, onDestroy } from 'svelte';
    
    export let imageUrl = '';
    
    let canvas;
    let locks = [];
    let lockImage = null;
    
    class LockSprite {
      constructor(canvas) {
        this.canvas = canvas;
        this.reset();
      }
  
      reset() {
        this.x = Math.random() * this.canvas.width;
        this.y = -50;
        this.speed = 1 + Math.random() * 3;
        this.size = 20 + Math.random() * 40;
        this.rotation = Math.random() * 360;
        this.rotationSpeed = -2 + Math.random() * 4;
      }
  
      update() {
        this.y += this.speed;
        this.rotation += this.rotationSpeed;
  
        if (this.y > this.canvas.height + 50) {
          this.reset();
        }
      }
  
      draw(ctx, image) {
        ctx.save();
        ctx.translate(this.x, this.y);
        ctx.rotate((this.rotation * Math.PI) / 180);
        
        if (image) {
          // Desenha a imagem se estiver disponível
          ctx.drawImage(
            image, 
            -this.size / 2, 
            -this.size / 2, 
            this.size, 
            this.size
          );
        } else {
          // Fallback: desenha um cadeado simples
          ctx.beginPath();
          ctx.strokeStyle = '#000';
          ctx.lineWidth = 2;
          
          const scale = this.size / 40;
          ctx.rect(-10 * scale, -5 * scale, 20 * scale, 25 * scale);
          ctx.moveTo(-7 * scale, -5 * scale);
          ctx.arc(0, -5 * scale, 7 * scale, Math.PI, 2 * Math.PI);
          
          ctx.stroke();
        }
        
        ctx.restore();
      }
    }
  
    let animationFrame;
    
    function resizeCanvas() {
      if (canvas) {
        canvas.width = window.innerWidth;
        canvas.height = window.innerHeight;
      }
    }
  
    onMount(() => {
      const ctx = canvas.getContext('2d');
      
      // Carrega a imagem se a URL foi fornecida
      if (imageUrl) {
        const img = new Image();
        img.onload = () => {
          lockImage = img;
        };
        img.src = imageUrl;
      }
  
      // Configura o canvas
      resizeCanvas();
      window.addEventListener('resize', resizeCanvas);
  
      // Inicializa os cadeados
      locks = Array.from({ length: 15 }, () => new LockSprite(canvas));
  
      // Função de animação
      function animate() {
        ctx.fillStyle = 'white';
        ctx.fillRect(0, 0, canvas.width, canvas.height);
  
        locks.forEach(lock => {
          lock.update();
          lock.draw(ctx, lockImage);
        });
  
        animationFrame = requestAnimationFrame(animate);
      }
  
      animate();
    });
  
    onDestroy(() => {
      if (animationFrame) {
        cancelAnimationFrame(animationFrame);
      }
      window.removeEventListener('resize', resizeCanvas);
    });
  </script>
  
  <canvas
    bind:this={canvas}
    style="position: fixed; top: 0; left: 0; width: 100%; height: 100%; z-index: -1;"
  />