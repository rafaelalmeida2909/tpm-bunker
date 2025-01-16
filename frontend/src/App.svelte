<script>
// @ts-nocheck

  import logo from "./assets/images/logo-universal.png";
  import FallingLocks from "./FallingLocks.svelte";
  import lock from "./assets/images/lock.png";
  import {
    InitializeDevice,
    IsDeviceInitialized,
    GetTPMStatus,
    GetDeviceInfo,
  } from "../wailsjs/go/main/App";

  let deviceInfo = null;
  let initialized = false;
  let status = null;
  let error = null;
  let loading = false;

  async function initializeTPM() {
    if (loading) return;
    loading = true;
    error = null;

    try {
      deviceInfo = await InitializeDevice();
      console.log("Device initialized:", deviceInfo);
    } catch (err) {
      error = `Falha na inicialização: ${err.message || err}`;
      console.error("Initialization error:", err);
    } finally {
      loading = false;
    }
  }

  async function checkInitialization() {
    if (loading) return;
    loading = true;
    error = null;

    try {
      initialized = await IsDeviceInitialized();
      console.log("Is initialized:", initialized);
    } catch (err) {
      error = `Falha ao verificar inicialização: ${err.message || err}`;
      console.error("Check initialization error:", err);
      initialized = false;
    } finally {
      loading = false;
    }
  }

  async function checkTPMStatus() {
    if (loading) return;
    loading = true;
    error = null;

    try {
      status = await GetTPMStatus();
      console.log("TPM Status:", status);
    } catch (err) {
      error = `Falha ao verificar status: ${err.message || err}`;
      console.error("Status check error:", err);
      status = null;
    } finally {
      loading = false;
    }
  }

  async function getDeviceDetails() {
    if (loading) return;
    loading = true;
    error = null;

    try {
      deviceInfo = await GetDeviceInfo();
      console.log("Device Info:", deviceInfo);
    } catch (err) {
      error = `Falha ao obter informações: ${err.message || err}`;
      console.error("Get info error:", err);
      deviceInfo = null;
    } finally {
      loading = false;
    }
  }

  async function runAllTests() {
    if (loading) return;
    error = null;

    try {
      await checkTPMStatus();
      await checkInitialization();

      if (!initialized) {
        await initializeTPM();
        await checkInitialization();
      }

      if (initialized) {
        await getDeviceDetails();
      }
    } catch (err) {
      error = `Erro nos testes: ${err.message || err}`;
      console.error("Test error:", err);
    }
  }
</script>

<main>
  <FallingLocks imageUrl={lock} />
  <img alt="Wails logo" id="logo" src={logo} />

  {#if error}
    <div class="error-box">
      <p>Error: {error}</p>
    </div>
  {/if}

  <div class="status-container">
    {#if status}
      <div class="status-box">
        <h3>TPM Status</h3>
        <p>Inicializado: {status.IsInitialized ? "Sim" : "Não"}</p>
        {#if status.HasError}
          <p class="error">Erro: {status.ErrorMessage}</p>
        {/if}
      </div>
    {/if}

    {#if deviceInfo}
      <div class="device-box">
        <h3>Informações do Dispositivo</h3>
        <p>UUID: {deviceInfo.UUID}</p>
        <p>Chave Pública: {deviceInfo.PublicKey}</p>
        <p>EK: {deviceInfo.EK}</p>
        <p>AIK: {deviceInfo.AIK}</p>
      </div>
    {/if}
  </div>

  <div class="button-container">
    <button class="btn" on:click={runAllTests}> Testar TPM </button>

    <button class="btn" on:click={initializeTPM} disabled={initialized}>
      Inicializar
    </button>

    <button class="btn" on:click={checkTPMStatus}> Verificar Status </button>
  </div>
</main>

<style>
  /* Para garantir que o conteúdo fique sobre a animação */
  main {
    position: relative;
    z-index: 1;
  }

  /* Resto dos seus estilos permanecem iguais */
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

  .error-box {
    background-color: #382e2e;
    border: 1px solid #ef5350;
    padding: 10px;
    margin: 10px 0;
    border-radius: 4px;
  }

  .status-container {
    display: flex;
    flex-direction: column;
    gap: 20px;
    margin: 20px 0;
  }

  .status-box,
  .device-box {
    background-color: #382e2e;
    padding: 15px;
    border-radius: 4px;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  }

  .button-container {
    display: flex;
    gap: 10px;
    justify-content: center;
    margin: 20px 0;
  }

  .btn {
    padding: 8px 16px;
    border-radius: 4px;
    border: none;
    background-color: #2196f3;
    color: white;
    cursor: pointer;
    transition: background-color 0.3s;
  }

  .btn:disabled {
    background-color: #bdbdbd;
    cursor: not-allowed;
  }

  .btn:hover:not(:disabled) {
    background-color: #1976d2;
  }
</style>
