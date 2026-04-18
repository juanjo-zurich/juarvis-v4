const checkboxes = document.querySelectorAll('.ide-card input[type="checkbox"]');
const installBtn = document.getElementById('install-btn');
const statusMessage = document.getElementById('status-message');
const consoleOutput = document.getElementById('console-output');
const consolePre = consoleOutput.querySelector('pre');

checkboxes.forEach(cb => {
    cb.addEventListener('change', () => {
        const anyChecked = Array.from(checkboxes).some(c => c.checked);
        installBtn.disabled = !anyChecked;
    });
});

installBtn.addEventListener('click', async () => {
    const targets = Array.from(checkboxes)
        .filter(c => c.checked)
        .map(c => c.value);

    installBtn.disabled = true;
    installBtn.innerText = 'Instalando...';
    statusMessage.classList.add('hidden');
    consoleOutput.classList.add('hidden');
    statusMessage.className = 'status-message'; // Reset

    try {
        const res = await fetch('/api/install', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ targets })
        });

        if (!res.ok) {
            let errorMsg = `Error ${res.status}`;
            let logs = "";
            try {
                const errData = await res.json();
                errorMsg = errData.error || errorMsg;
                logs = errData.logs || "";
            } catch(e) {}
            
            statusMessage.classList.remove('hidden');
            statusMessage.classList.add('error');
            statusMessage.innerText = 'Fallo al instalar: ' + errorMsg;
            if (logs) {
                consoleOutput.classList.remove('hidden');
                consolePre.innerText = logs;
            }
            throw new Error(errorMsg);
        }

        const data = await res.json();
        
        statusMessage.classList.remove('hidden');
        statusMessage.classList.add('success');
        statusMessage.innerText = '¡Ecosistema distribuido con éxito! Revisa la consola inferior y verifica tu editor.';
        
        if (data.logs) {
            consoleOutput.classList.remove('hidden');
            consolePre.innerText = data.logs;
        }

        installBtn.innerText = '¡Completado!';
        updateStatus();
    } catch (e) {
        if (!statusMessage.classList.contains('error')) {
            statusMessage.classList.remove('hidden');
            statusMessage.classList.add('error');
            statusMessage.innerText = 'Error de conexión o fallo no controlado.';
        }
        installBtn.disabled = false;
        installBtn.innerText = 'Reintentar Instalación';
    }
});

async function updateStatus() {
    try {
        const res = await fetch('/api/status');
        if (res.ok) {
            const data = await res.json();
            document.getElementById('stat-snapshots').innerText = data.snapshots;
            // Calcular uptime aproximado
            const startTime = new Date(data.time).getTime();
            const now = new Date().getTime();
            const mins = Math.floor((now - startTime) / 60000);
            document.getElementById('stat-uptime').innerText = mins + 'm';
        }
    } catch (e) {}
}

setInterval(updateStatus, 5000);
updateStatus();
