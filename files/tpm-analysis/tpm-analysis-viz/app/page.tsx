'use client';

import { useEffect, useRef, useState } from 'react';
import { Bar, BarChart, CartesianGrid, Legend, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts';

export default function TPMBarGraphs() {
  const [data, setData] = useState({});
  // Refs for each chart container
  const chartRefs = useRef({});

  const processFileContent = (content) => {
    const lines = content.split('\n');
    const result = { encryption: {}, decryption: {} };
    let currentSection = null;

    lines.forEach((line) => {
      if (line.includes('Encriptação:') || line.includes('Encryption')) {
        currentSection = 'encryption';
      } else if (line.includes('Decriptação:') || line.includes('Decryption')) {
        currentSection = 'decryption';
      }

      const statsRegex = /Size:\s*([\d.]+)\s*MB.*?Duration:\s*([\d.]+)\s*(ms|s)/i;
      const matches = line.match(statsRegex);

      if (matches) {
        const size = parseFloat(matches[1]);
        let duration = parseFloat(matches[2]);
        if (matches[3].toLowerCase() === 's') duration *= 1000;

        if (!result[currentSection][size]) result[currentSection][size] = [];
        result[currentSection][size].push(duration);
      }
    });

    return Object.keys(result).reduce((acc, key) => {
      acc[key] = Object.entries(result[key]).map(([size, durations]) => ({
        size: parseFloat(size),
        duration: durations.reduce((a, b) => a + b, 0) / durations.length
      })).sort((a, b) => a.size - b.size);
      return acc;
    }, {});
  };

  useEffect(() => {
    const loadData = async () => {
      const machines = ['machine1', 'machine2', 'machine3'];
      const results = {};

      for (const machine of machines) {
        results[machine] = {};
        try {
          const withTPMResponse = await fetch(`/${machine}/result_with_tpm.txt`);
          const withoutTPMResponse = await fetch(`/${machine}/result_without_tpm.txt`);
          if (!withTPMResponse.ok || !withoutTPMResponse.ok) throw new Error('Error fetching files');

          const withTPMData = processFileContent(await withTPMResponse.text());
          const withoutTPMData = processFileContent(await withoutTPMResponse.text());

          results[machine] = { withTPM: withTPMData, withoutTPM: withoutTPMData };
        } catch (error) {
          console.error(`Error loading data for ${machine}:`, error);
        }
      }
      setData(results);
    };

    loadData();
  }, []);

  // Load html2canvas dynamically when needed
  const exportChart = async (chartId, filename) => {
    // Dynamically import html2canvas
    const html2canvas = (await import('html2canvas')).default;
    
    const element = chartRefs.current[chartId];
    if (!element) {
      console.error(`Chart element with ID ${chartId} not found`);
      return;
    }

    try {
      // Create a temporary container with exact dimensions and white background
      const tempContainer = document.createElement('div');
      tempContainer.style.width = '558px';
      tempContainer.style.height = '228px';
      tempContainer.style.backgroundColor = 'white';
      tempContainer.style.position = 'absolute';
      tempContainer.style.left = '-9999px';
      tempContainer.style.top = '-9999px';
      document.body.appendChild(tempContainer);
      
      // Clone the chart element
      const clone = element.cloneNode(true);
      clone.style.width = '558px';
      clone.style.height = '228px';
      
      // Remove the title if it exists (assuming first child might be title container)
      const titleContainer = clone.querySelector('div.flex.justify-between.items-center');
      if (titleContainer) {
        clone.removeChild(titleContainer);
      }
      
      tempContainer.appendChild(clone);
      
      // Render the chart to canvas
      const canvas = await html2canvas(tempContainer, {
        backgroundColor: 'white',
        width: 558,
        height: 228,
        scale: 1,
        logging: false,
        useCORS: true
      });
      
      // Clean up
      document.body.removeChild(tempContainer);
      
      // Create download link
      const link = document.createElement('a');
      link.download = filename;
      link.href = canvas.toDataURL('image/png');
      link.click();
    } catch (error) {
      console.error('Failed to export chart:', error);
    }
  };
  
  // Function to render encryption chart for a machine with Bar chart
  const renderEncryptionChart = (machineData, machineName, machineKey) => {
    if (!machineData.withTPM || !machineData.withoutTPM) return null;
    
    const allSizes = [...new Set([
      ...machineData.withTPM.encryption.map(d => d.size),
      ...machineData.withoutTPM.encryption.map(d => d.size)
    ])].sort((a, b) => a - b);

    // Transform data for bar chart
    const chartData = allSizes.map(size => {
      const withTPM = machineData.withTPM.encryption.find(d => d.size === size)?.duration || 0;
      const withoutTPM = machineData.withoutTPM.encryption.find(d => d.size === size)?.duration || 0;
      
      return {
        size: `${size} MB`,
        'With TPM': withTPM,
        'Without TPM': withoutTPM
      };
    });

    const chartId = `${machineKey}_encryption`;

    return (
      <div className="w-full mb-8 p-4 bg-white rounded-lg shadow export-container">
        <div className="flex justify-between items-center mb-4">
          <h2 className="text-xl font-bold">{machineName} - Encryption Performance Comparison</h2>
          <button 
            onClick={() => exportChart(chartId, `${machineKey.replace('machine', 'm')}_cripto_bar.png`)}
            className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 transition-colors"
          >
            Export Image
          </button>
        </div>
        <div 
          className="h-[400px] bg-white" 
          ref={el => chartRefs.current[chartId] = el}
        >
          <ResponsiveContainer width="100%" height="100%">
            <BarChart 
              data={chartData}
              margin={{ top: 20, right: 30, left: 20, bottom: 30 }}
            >
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis 
                dataKey="size" 
                label={{ value: 'Size (MB)', position: 'insideBottom', offset: -5, fill: '#000' }} 
                tick={{ fill: '#000' }}
              />
              <YAxis 
                label={{ value: 'Duration (ms)', angle: -90, position: 'insideLeft', offset: 5, fill: '#000' }} 
                tick={{ fill: '#000' }} 
              />
              <Tooltip formatter={(value) => [`${value.toFixed(2)}ms`, null]} />
              <Legend align="left" verticalAlign="top" height={36} />
              <Bar dataKey="With TPM" fill="#286be6" barSize={40} />
              <Bar dataKey="Without TPM" fill="#23a353" barSize={40} />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </div>
    );
  };

  // Function to render decryption chart for a machine
  const renderDecryptionChart = (machineData, machineName, machineKey) => {
    if (!machineData.withTPM || !machineData.withoutTPM) return null;
    
    const allSizes = [...new Set([
      ...machineData.withTPM.decryption.map(d => d.size),
      ...machineData.withoutTPM.decryption.map(d => d.size)
    ])].sort((a, b) => a - b);

    // Transform data for bar chart
    const chartData = allSizes.map(size => {
      const withTPM = machineData.withTPM.decryption.find(d => d.size === size)?.duration || 0;
      const withoutTPM = machineData.withoutTPM.decryption.find(d => d.size === size)?.duration || 0;
      
      return {
        size: `${size} MB`,
        'With TPM': withTPM,
        'Without TPM': withoutTPM
      };
    });

    const chartId = `${machineKey}_decryption`;

    return (
      <div className="w-full mb-8 p-4 bg-white rounded-lg shadow">
        <div className="flex justify-between items-center mb-4">
          <h2 className="text-xl font-bold">{machineName} - Decryption Performance Comparison</h2>
          <button 
            onClick={() => exportChart(chartId, `${machineKey.replace('machine', 'm')}_descripto_bar.png`)}
            className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 transition-colors"
          >
            Export Image
          </button>
        </div>
        <div 
          className="h-[400px] bg-white"
          ref={el => chartRefs.current[chartId] = el}
        >
          <ResponsiveContainer width="100%" height="100%">
            <BarChart 
              data={chartData}
              margin={{ top: 20, right: 30, left: 20, bottom: 30 }}
            >
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis 
                dataKey="size" 
                label={{ value: 'Size (MB)', position: 'insideBottom', offset: -5, fill: '#000' }} 
                tick={{ fill: '#000' }}
              />
              <YAxis 
                label={{ value: 'Duration (ms)', angle: -90, position: 'insideLeft', offset: 5, fill: '#000' }} 
                tick={{ fill: '#000' }} 
              />
              <Tooltip formatter={(value) => [`${value.toFixed(2)}ms`, null]} />
              <Legend align="left" verticalAlign="top" height={36} />
              <Bar dataKey="With TPM" fill="#286be6" barSize={40} />
              <Bar dataKey="Without TPM" fill="#23a353" barSize={40} />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </div>
    );
  };

  return (
    <div className="p-6 bg-gray-50 min-h-screen">
      <h1 className="text-2xl font-bold mb-6">TPM Performance Comparison</h1>
      {Object.entries(data).map(([machine, machineData]) => (
        <div key={machine} className="mb-10">
          <h2 className="text-xl font-bold mb-4">{machine.charAt(0).toUpperCase() + machine.slice(1)}</h2>
          {renderEncryptionChart(machineData, machine.charAt(0).toUpperCase() + machine.slice(1), machine)}
          {renderDecryptionChart(machineData, machine.charAt(0).toUpperCase() + machine.slice(1), machine)}
        </div>
      ))}
      <div className="mt-4 p-4 bg-gray-100 rounded">
        <h3 className="font-bold">Debug Info:</h3>
        <pre className="mt-2 text-sm overflow-auto">{JSON.stringify(data, null, 2)}</pre>
      </div>
    </div>
  );
}