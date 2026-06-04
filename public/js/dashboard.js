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

// Reusable Copy Text function
function copyText(text) {
    // 检查 Clipboard API 是否可用
    if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(text).then(() => {
            showToast('已复制到剪贴板');
        }).catch(err => {
            console.error('Clipboard API 失败:', err);
            fallbackCopy(text);
        });
    } else {
        // 降级方案
        fallbackCopy(text);
    }
}

// 降级方案：使用 textarea 和 execCommand
function fallbackCopy(text) {
    const textarea = document.createElement('textarea');
    textarea.value = text;
    textarea.style.position = 'fixed';
    textarea.style.top = '-9999px';
    textarea.style.left = '-9999px';
    textarea.style.opacity = '0';
    document.body.appendChild(textarea);

    textarea.select();
    textarea.setSelectionRange(0, text.length);

    try {
        const successful = document.execCommand('copy');
        if (successful) {
            showToast('已复制到剪贴板');
        } else {
            showToast('复制失败，请手动复制', false);
        }
    } catch (err) {
        console.error('execCommand 复制失败:', err);
        showToast('复制失败，请手动复制', false);
    }

    document.body.removeChild(textarea);
}

// Client-side Table Paginator Class
class TablePaginator {
    constructor(table, pageSize = 10) {
        this.table = table;
        this.tbody = table.querySelector('tbody');
        this.pageSize = pageSize;
        this.currentPage = 1;

        // Find or create container for pagination controls
        this.card = table.closest('.table-card');
        if (!this.card) return;

        this.controlsContainer = document.createElement('div');
        this.controlsContainer.className = 'pagination-controls';
        this.card.appendChild(this.controlsContainer);

        table.paginator = this; // Bind to DOM element
        this.update();
    }

    getVisibleRows() {
        return Array.from(this.tbody.querySelectorAll('tr')).filter(row => {
            if (row.cells.length === 1 && row.cells[0].getAttribute('colspan')) {
                return false; // Skip no data placeholder
            }
            return row.style.display !== 'none';
        });
    }

    update() {
        const visibleRows = this.getVisibleRows();
        const totalRows = visibleRows.length;
        const totalPages = Math.ceil(totalRows / this.pageSize) || 1;

        if (this.currentPage > totalPages) {
            this.currentPage = totalPages;
        }

        const allRows = Array.from(this.tbody.querySelectorAll('tr')).filter(row => {
            return !(row.cells.length === 1 && row.cells[0].getAttribute('colspan'));
        });

        // Hide all rows initially
        allRows.forEach(row => row.classList.add('paginated-hidden'));

        // Show visible rows on current page
        const start = (this.currentPage - 1) * this.pageSize;
        const end = start + this.pageSize;

        visibleRows.forEach((row, index) => {
            if (index >= start && index < end) {
                row.classList.remove('paginated-hidden');
            }
        });

        this.renderControls(totalPages, totalRows);
    }

    renderControls(totalPages, totalRows) {
        if (totalRows <= this.pageSize && this.currentPage === 1) {
            this.controlsContainer.style.display = 'none';
            return;
        }
        this.controlsContainer.style.display = 'flex';
        this.controlsContainer.innerHTML = '';

        const info = document.createElement('div');
        info.className = 'pagination-info';
        const startIdx = totalRows === 0 ? 0 : (this.currentPage - 1) * this.pageSize + 1;
        const endIdx = Math.min(this.currentPage * this.pageSize, totalRows);
        info.textContent = `显示 ${startIdx}-${endIdx} 条，共 ${totalRows} 条`;
        this.controlsContainer.appendChild(info);

        const btnGroup = document.createElement('div');
        btnGroup.className = 'pagination-buttons';

        // Prev button
        const prevBtn = document.createElement('button');
        prevBtn.className = 'btn-pagination';
        prevBtn.innerHTML = '<i class="fas fa-chevron-left"></i>';
        prevBtn.disabled = this.currentPage === 1;
        prevBtn.onclick = (e) => {
            e.preventDefault();
            if (this.currentPage > 1) {
                this.currentPage--;
                this.update();
            }
        };
        btnGroup.appendChild(prevBtn);

        // Page numbers
        for (let i = 1; i <= totalPages; i++) {
            if (totalPages > 5) {
                if (i !== 1 && i !== totalPages && Math.abs(i - this.currentPage) > 1) {
                    if (i === 2 || i === totalPages - 1) {
                        const dots = document.createElement('span');
                        dots.className = 'pagination-dots';
                        dots.textContent = '...';
                        btnGroup.appendChild(dots);
                    }
                    continue;
                }
            }

            const pageBtn = document.createElement('button');
            pageBtn.className = `btn-pagination ${i === this.currentPage ? 'active' : ''}`;
            pageBtn.textContent = i;
            pageBtn.onclick = (e) => {
                e.preventDefault();
                this.currentPage = i;
                this.update();
            };
            btnGroup.appendChild(pageBtn);
        }

        // Next button
        const nextBtn = document.createElement('button');
        nextBtn.className = 'btn-pagination';
        nextBtn.innerHTML = '<i class="fas fa-chevron-right"></i>';
        nextBtn.disabled = this.currentPage === totalPages;
        nextBtn.onclick = (e) => {
            e.preventDefault();
            if (this.currentPage < totalPages) {
                this.currentPage++;
                this.update();
            }
        };
        btnGroup.appendChild(nextBtn);

        this.controlsContainer.appendChild(btnGroup);
    }
}

// Table item beautification (platforms, click-to-copy tokens)
function beautifyTableItems() {
    // 1. Truncate long tokens and add copy action
    document.querySelectorAll('.token-list-hash').forEach(el => {
        const fullToken = el.getAttribute('title') || el.textContent.trim();

        if (fullToken.length > 20) {
            el.textContent = fullToken.substring(0, 8) + '...' + fullToken.substring(fullToken.length - 8);
        }
        el.innerHTML += "  " + "<i class='fa-regular fa-copy'></i>";

        el.addEventListener('click', (e) => {
            e.preventDefault();
            e.stopPropagation();
            copyText(fullToken);
        });
    });

    // 2. Transform raw platform text into beautiful badges with icons
    document.querySelectorAll('td, .badge').forEach(el => {
        // Skip if element already has complex markup
        if (el.children.length > 0 && !el.classList.contains('badge')) return;

        const txt = el.textContent.trim().toLowerCase();
        if (['android', 'ios', 'windows', 'linux', 'mac'].includes(txt)) {
            let iconClass = '';
            let label = txt;
            switch (txt) {
                case 'android':
                    iconClass = 'fab fa-android';
                    break;
                case 'ios':
                    iconClass = 'fab fa-apple';
                    label = 'iOS';
                    break;
                case 'windows':
                    iconClass = 'fab fa-windows';
                    label = 'Windows';
                    break;
                case 'linux':
                    iconClass = 'fab fa-linux';
                    label = 'Linux';
                    break;
                case 'mac':
                    iconClass = 'fab fa-apple';
                    label = 'macOS';
                    break;
            }

            el.className = `badge badge-platform badge-${txt}`;
            el.innerHTML = `<i class="${iconClass}"></i> ${label}`;
        }
    });
}

// Unified client-side table filter
function initTableSearch(inputId, tbodyId) {
    const input = document.getElementById(inputId);
    const tbody = document.getElementById(tbodyId);
    if (!input || !tbody) return;

    input.addEventListener('input', () => {
        const val = input.value.toLowerCase().trim();
        const rows = tbody.querySelectorAll('tr');

        rows.forEach(row => {
            if (row.cells.length === 1 && row.cells[0].getAttribute('colspan')) return;

            let text = '';
            Array.from(row.cells).forEach(cell => {
                // Ignore operation/action cell text to avoid false matches on buttons
                if (cell.classList.contains('actions-cell') || cell.querySelector('form') || cell.querySelector('button')) return;
                text += ' ' + cell.textContent.toLowerCase();
            });

            const match = text.includes(val);
            row.style.display = match ? '' : 'none';
        });

        const table = tbody.closest('table');
        if (table && table.paginator) {
            table.paginator.currentPage = 1;
            table.paginator.update();
        }
    });
}

// Initialize tables on page load
document.addEventListener('DOMContentLoaded', () => {
    // 1. Run item beautifier
    beautifyTableItems();

    // 2. Setup paginator on all tables in table cards
    document.querySelectorAll('.table-card table').forEach(table => {
        const rows = table.querySelectorAll('tbody tr');
        // Only paginate if table actually has records and more than 10 items
        if (rows.length > 0 && !(rows.length === 1 && rows[0].cells.length === 1 && rows[0].cells[0].getAttribute('colspan'))) {
            new TablePaginator(table, 10);
        }
    });

    // 3. Connect existing filter-inputs
    const searchInput = document.getElementById('search-input');
    if (searchInput) {
        // If there's an inline search handler (like filterApps), we run the paginator check after it
        searchInput.addEventListener('input', () => {
            const table = document.querySelector('.table-card table');
            if (table && table.paginator) {
                table.paginator.currentPage = 1;
                setTimeout(() => table.paginator.update(), 0);
            }
        });
    }
});

