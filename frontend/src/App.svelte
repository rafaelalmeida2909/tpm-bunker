<script>
  import { GetTPMStatus } from '../wailsjs/go/main/App.js';
  import logo from './assets/images/logo-universal.png';

  let status = null;
  let error = null;

  async function checkTPMStatus() {
    try {
      status = await GetTPMStatus()
      error = null
    } catch (err) {
      error = err
      status = null
    }
  }
</script>

<main>
  <img alt="Wails logo" id="logo" src="{logo}">
  
  {#if error}
    <div class="error">Error: {error}</div>
  {:else if status}
    <div class="status">
      <p>Available: {status.available ? 'Yes' : 'No'}</p>
      <p>Version: {status.version}</p>
      {#if status.error}
        <p class="error">TPM Error: {status.error}</p>
      {/if}
    </div>
  {:else}
    <p>Click the button to check TPM status</p>
  {/if}

  <div class="input-box">
    <button class="btn" on:click={checkTPMStatus}>Check</button>
  </div>
</main>

<style>

  #logo {
    display: block;
    width: 50%;
    height: 50%;
    margin: auto;
    padding: 10% 0 0;
    background-position: center;
    background-repeat: no-repeat;
    background-size: 100% 100%;
    background-origin: content-box;
  }

  .input-box .btn {
    width: 60px;
    height: 30px;
    line-height: 30px;
    border-radius: 3px;
    border: none;
    margin: 0 0 0 20px;
    padding: 0 8px;
    cursor: pointer;
  }

  .input-box .btn:hover {
    background-image: linear-gradient(to top, #cfd9df 0%, #e2ebf0 100%);
    color: #333333;
  }

</style>
