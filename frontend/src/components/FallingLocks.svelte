<script>
  import { onDestroy } from "svelte";
  import Lock from "svelte-icons/fa/FaLock.svelte";
  export let count = 15;

  let locks = [];
  let animationFrame;

  function createLocks() {
    locks = Array(count)
      .fill(0)
      .map((_, i) => ({
        id: i,
        left: Math.random() * 100, 
        delay: Math.random() * 0.3, 
        duration: 1 + Math.random() * 1.5,
        size: Math.random() * 10 + 15,
      }));
  }

  createLocks();

  onDestroy(() => {
    if (animationFrame) {
      cancelAnimationFrame(animationFrame);
    }
  });
</script>

<div class="fixed inset-0 pointer-events-none z-50">
  {#each locks as lock (lock.id)}
    <div
      class="absolute text-blue-500 opacity-50 animate-fall will-change-transform"
      style="left: {lock.left}%; animation-delay: {lock.delay}s; animation-duration: {lock.duration}s"
    >
      <div class="w-4 h-4" style="width: {lock.size}px; height: {lock.size}px">
        <Lock />
      </div>
    </div>
  {/each}
</div>

<style>
  @keyframes fall {
    0% {
      transform: translateY(-20px) rotate(0deg);
      opacity: 0.8;
    }
    100% {
      transform: translateY(100vh) rotate(360deg);
      opacity: 0;
    }
  }
  .animate-fall {
    animation: fall 1.2s linear forwards;
  }
</style>
