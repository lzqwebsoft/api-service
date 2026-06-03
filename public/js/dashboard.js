// Modal Helpers
function openModal(id) {
    const modal = document.getElementById(id);
    if (modal) {
        modal.classList.add('active');
    }
}

function closeModal(id) {
    const modal = document.getElementById(id);
    if (modal) {
        modal.classList.remove('active');
    }
}

// Copy generated token to clipboard
function copyGeneratedToken() {
    const tokenEl = document.getElementById('generated-token-text');
    if (!tokenEl) return;
    
    const token = tokenEl.innerText;
    navigator.clipboard.writeText(token).then(() => {
        showToast('Token 已复制到剪贴板');
    }).catch(err => {
        showToast('复制失败，请手动选择复制', false);
    });
}

// Local client-side table filter
function filterApps() {
    const input = document.getElementById('search-input');
    if (!input) return;
    
    const val = input.value.toLowerCase().trim();
    const rows = document.querySelectorAll('#app-table-body tr');
    
    rows.forEach(row => {
        const titleEl = row.querySelector('.app-name-title');
        const idEl = row.querySelector('.app-name-id');
        
        if (!titleEl || !idEl) return;
        
        const match = titleEl.textContent.toLowerCase().includes(val) || 
                      idEl.textContent.toLowerCase().includes(val);
                      
        row.style.display = match ? '' : 'none';
    });
}

// Client-side Toast notification helper
function showToast(message, isSuccess = true) {
    const toast = document.getElementById('toast-notify');
    const icon = document.getElementById('toast-icon');
    const text = document.getElementById('toast-text');
    
    if (!toast || !icon || !text) return;

    text.innerText = message;
    
    // Reset styles
    toast.className = 'toast';
    
    if (isSuccess) {
        toast.classList.add('toast-success');
        icon.className = 'fas fa-check-circle';
        icon.style.color = 'var(--success)';
    } else {
        toast.classList.add('toast-error');
        icon.className = 'fas fa-times-circle';
        icon.style.color = 'var(--danger)';
    }
    
    toast.classList.add('active');
    
    setTimeout(() => {
        toast.classList.remove('active');
    }, 3000);
}
