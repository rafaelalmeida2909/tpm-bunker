<!-- FallingLocks.svelte -->
<script lang="ts">
  import { onDestroy, onMount } from 'svelte';
  import lockImageUrl from './assets/images/lock.png';
  
  let canvas: HTMLCanvasElement;
  let locks: LockSprite[] = [];
  let lockImage: HTMLImageElement | null = null;
  let animationFrame: number;
  
  class LockSprite {
    private canvas: HTMLCanvasElement;
    private x: number;
    private y: number;
    private speed: number;
    private size: number;
    private rotation: number;
    private rotationSpeed: number;
  
    constructor(canvas: HTMLCanvasElement) {
      this.canvas = canvas;
      this.x = 0;
      this.y = 0;
      this.speed = 0;
      this.size = 0;
      this.rotation = 0;
      this.rotationSpeed = 0;
      this.reset();
    }
  
    reset(): void {
      this.x = Math.random() * this.canvas.width;
      this.y = -50;
      this.speed = 1 + Math.random() * 3;
      this.size = 20 + Math.random() * 40;
      this.rotation = Math.random() * 360;
      this.rotationSpeed = -2 + Math.random() * 4;
    }
  
    update(): void {
      this.y += this.speed;
      this.rotation += this.rotationSpeed;
  
      if (this.y > this.canvas.height + 50) {
        this.reset();
      }
    }
  
    draw(ctx: CanvasRenderingContext2D, image: HTMLImageElement | null): void {
      if (!ctx) return;
      
      ctx.save();
      ctx.translate(this.x, this.y);
      ctx.rotate((this.rotation * Math.PI) / 180);
      
      if (image) {
        ctx.drawImage(
          image, 
          -this.size / 2, 
          -this.size / 2, 
          this.size, 
          this.size
        );
      } else {
        // Fallback: draw a simple lock shape
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
  
  function resizeCanvas(): void {
    if (canvas) {
      canvas.width = window.innerWidth;
      canvas.height = window.innerHeight;
    }
  }
  
  onMount(() => {
    // Ensure canvas is available
    if (!canvas) {
      console.error('Canvas element not found');
      return;
    }
    
    const ctx = canvas.getContext('2d');
    if (!ctx) {
      console.error('Unable to get 2D rendering context');
      return;
    }
    
    // Load the lock image
    const img = new Image();
    img.onload = () => {
      lockImage = img;
    };
    img.onerror = () => {
      console.warn('Failed to load lock image');
    };
    img.src = lockImageUrl;
  
    // Setup canvas
    resizeCanvas();
    window.addEventListener('resize', resizeCanvas);
  
    // Initialize locks
    locks = Array.from({ length: 15 }, () => new LockSprite(canvas));
  
    // Animation function
    function animate(): void {
      // Additional null checks
      if (!ctx || !canvas) return;
      
      ctx.fillStyle = 'white';
      ctx.fillRect(0, 0, canvas.width, canvas.height);
  
      locks.forEach(lock => {
        lock.update();
        lock.draw(ctx, lockImage);
      });
  
      animationFrame = requestAnimationFrame(animate);
    }
  
    animate();
  
    // Return a cleanup function
    return () => {
      if (animationFrame) {
        cancelAnimationFrame(animationFrame);
      }
      window.removeEventListener('resize', resizeCanvas);
    };
  });
  
  // Note: onDestroy is not strictly necessary with the cleanup function in onMount
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