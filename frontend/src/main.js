// 全局变量存储后端绑定
let backend = null;

// 页面切换函数 - 暴露到全局作用域
window.showHomePage = function() {
    document.getElementById('homePage').style.display = 'flex';
    document.getElementById('sendPage').style.display = 'none';
    document.getElementById('receivePage').style.display = 'none';
}

window.showSendPage = function() {
    document.getElementById('homePage').style.display = 'none';
    document.getElementById('sendPage').style.display = 'flex';
    document.getElementById('receivePage').style.display = 'none';
    document.getElementById('sendStatus').textContent = '就绪';
}

window.showReceivePage = async function() {
    document.getElementById('homePage').style.display = 'none';
    document.getElementById('sendPage').style.display = 'none';
    document.getElementById('receivePage').style.display = 'flex';
    document.getElementById('receiveStatus').textContent = '正在启动接收...';
    
    // 自动开始接收
    if (!await initBackend()) {
        document.getElementById('receiveStatus').textContent = '后端未就绪';
        return;
    }
    
    try {
        document.getElementById('receiveStatus').textContent = '正在接收...';
        await backend.Receive();
    } catch (error) {
        console.error('接收失败:', error);
        document.getElementById('receiveStatus').textContent = '接收失败: ' + error;
    }
}

// 初始化后端绑定
async function initBackend() {
    if (window.go && window.go.main && window.go.main.App) {
        backend = window.go.main.App;
        return true;
    }
    return false;
}

// 发送文件
window.sendFile = async function() {
    if (!await initBackend()) {
        document.getElementById('sendStatus').textContent = '后端未就绪';
        return;
    }
    
    const selectedPath = window.currentSelectedPath;
    if (!selectedPath) {
        document.getElementById('sendStatus').textContent = '请选择要发送的文件或文件夹';
        return;
    }
    
    try {
        document.getElementById('sendStatus').textContent = '正在发送...';
        await backend.Send(selectedPath);
    } catch (error) {
        console.error('发送失败:', error);
        document.getElementById('sendStatus').textContent = '发送失败: ' + error;
    }
}

// 接收文件
window.receiveFile = async function() {
    if (!await initBackend()) {
        document.getElementById('receiveStatus').textContent = '后端未就绪';
        return;
    }
    
    try {
        document.getElementById('receiveStatus').textContent = '正在接收...';
        await backend.Receive();
    } catch (error) {
        console.error('接收失败:', error);
        document.getElementById('receiveStatus').textContent = '接收失败: ' + error;
    }
}


// 重置发送状态
window.resetSendState = async function() {
    if (!await initBackend()) {
        return;
    }
    
    try {
        document.getElementById('sendPath').value = '';
        document.getElementById('sendStatus').textContent = '就绪';
        resetProgressBars();
    } catch (error) {
        console.error('重置发送状态失败:', error);
    }
}

// 重置接收状态并重启接收模式
window.resetReceiveState = async function() {
    if (!await initBackend()) {
        return;
    }
    
    try {
        document.getElementById('receiveStatus').textContent = '正在重启接收...';
        
        // 重启接收模式
        setTimeout(async () => {
            try {
                await backend.RestartReceive();
                document.getElementById('receiveStatus').textContent = '正在接收...';
            } catch (error) {
                console.error('重启接收模式失败:', error);
                document.getElementById('receiveStatus').textContent = '正在接收...: ' + error;
            }
        }, 500);
    } catch (error) {
        console.error('重置接收状态失败:', error);
    }
}

// 更新发送状态
function updateSendStatus(status) {
    document.getElementById('sendStatus').textContent = status;
}

// 更新接收状态
function updateReceiveStatus(status) {
    document.getElementById('receiveStatus').textContent = status;
}

// 更新进度条
function updateProgressBar(stats) {
    const progress = Math.min(100, Math.max(0, stats.progress || 0));
    const progressPercent = progress.toFixed(1) + '%';
    
    // 更新发送页面进度
    const sendProgressBar = document.getElementById('sendProgressBar');
    const sendProgressPercent = document.getElementById('sendProgressPercent');
    const sendProgressSpeed = document.getElementById('sendProgressSpeed');
    const sendProgressETA = document.getElementById('sendProgressETA');
    
    if (sendProgressBar && sendProgressPercent) {
        sendProgressBar.style.width = progressPercent;
        sendProgressPercent.textContent = progressPercent;
        
        // 更新速度显示 - 使用后端提供的CurrentSpeed字段
        if (sendProgressSpeed && stats.currentSpeed !== undefined) {
            const speedMB = stats.currentSpeed.toFixed(1);
            sendProgressSpeed.textContent = `${speedMB} MB/s`;
        } else if (sendProgressSpeed) {
            sendProgressSpeed.textContent = '0 MB/s';
        }
        
        // 更新剩余时间显示 - 使用后端提供的EstimatedTime字段
        if (sendProgressETA && stats.estimatedTime) {
            sendProgressETA.textContent = stats.estimatedTime;
        } else if (sendProgressETA) {
            sendProgressETA.textContent = '计算中...';
        }
    }
    
    // 更新接收页面进度
    const receiveProgressBar = document.getElementById('receiveProgressBar');
    const receiveProgressPercent = document.getElementById('receiveProgressPercent');
    const receiveProgressSpeed = document.getElementById('receiveProgressSpeed');
    const receiveProgressETA = document.getElementById('receiveProgressETA');
    
    if (receiveProgressBar && receiveProgressPercent) {
        receiveProgressBar.style.width = progressPercent;
        receiveProgressPercent.textContent = progressPercent;
        
        // 更新速度显示 - 使用后端提供的CurrentSpeed字段
        if (receiveProgressSpeed && stats.currentSpeed !== undefined) {
            const speedMB = stats.currentSpeed.toFixed(1);
            receiveProgressSpeed.textContent = `${speedMB} MB/s`;
        } else if (receiveProgressSpeed) {
            receiveProgressSpeed.textContent = '0 MB/s';
        }
        
        // 更新剩余时间显示 - 使用后端提供的EstimatedTime字段
        if (receiveProgressETA && stats.estimatedTime) {
            receiveProgressETA.textContent = stats.estimatedTime;
        } else if (receiveProgressETA) {
            receiveProgressETA.textContent = '计算中...';
        }
    }
}

// 重置进度条
function resetProgressBars() {
    // 重置发送页面进度
    const sendProgressBar = document.getElementById('sendProgressBar');
    const sendProgressPercent = document.getElementById('sendProgressPercent');
    const sendProgressSpeed = document.getElementById('sendProgressSpeed');
    const sendProgressETA = document.getElementById('sendProgressETA');
    
    if (sendProgressBar && sendProgressPercent) {
        sendProgressBar.style.width = '0%';
        sendProgressPercent.textContent = '0%';
    }
    if (sendProgressSpeed) {
        sendProgressSpeed.textContent = '0 MB/s';
    }
    if (sendProgressETA) {
        sendProgressETA.textContent = '计算中...';
    }
    
    // 重置接收页面进度
    const receiveProgressBar = document.getElementById('receiveProgressBar');
    const receiveProgressPercent = document.getElementById('receiveProgressPercent');
    const receiveProgressSpeed = document.getElementById('receiveProgressSpeed');
    const receiveProgressETA = document.getElementById('receiveProgressETA');
    
    if (receiveProgressBar && receiveProgressPercent) {
        receiveProgressBar.style.width = '0%';
        receiveProgressPercent.textContent = '0%';
    }
    if (receiveProgressSpeed) {
        receiveProgressSpeed.textContent = '0 MB/s';
    }
    if (receiveProgressETA) {
        receiveProgressETA.textContent = '计算中...';
    }
}

// 监听后端事件
if (window.runtime && window.runtime.EventsOn) {
    window.runtime.EventsOn('status-updated', (status) => {
        // 根据当前页面更新对应的状态
        if (document.getElementById('sendPage').style.display === 'flex') {
            updateSendStatus(status);
        } else if (document.getElementById('receivePage').style.display === 'flex') {
            updateReceiveStatus(status);
        }
    });

    window.runtime.EventsOn('operation-completed', () => {
        // 根据当前页面更新对应的状态
        if (document.getElementById('sendPage').style.display === 'flex') {
            document.getElementById('sendStatus').textContent = '操作完成';
        } else if (document.getElementById('receivePage').style.display === 'flex') {
            document.getElementById('receiveStatus').textContent = '操作完成';
        }
    });

    window.runtime.EventsOn('stats-updated', (stats) => {
        // 更新进度条
        updateProgressBar(stats);
    });

}

// 文件选择功能实现
function setupFileSelection() {
    const dropZone = document.getElementById('dropZone');
    const selectedFiles = document.getElementById('selectedFiles');
    const fileList = document.getElementById('fileList');
    
    if (!dropZone) return;
    
    // 创建文件选择对话框
    function createFileSelectionDialog() {
        const dialog = document.createElement('div');
        dialog.className = 'file-selection-dialog';
        dialog.innerHTML = `
            <div class="dialog-overlay"></div>
            <div class="dialog-content">
                <div class="dialog-header">
                    <h3>选择文件或文件夹</h3>
                    <button class="dialog-close">&times;</button>
                </div>
                <div class="dialog-body">
                    <button class="selection-button file-button">
                        <span class="button-icon">📄</span>
                        <span class="button-text">选择文件</span>
                    </button>
                    <button class="selection-button folder-button">
                        <span class="button-icon">📁</span>
                        <span class="button-text">选择文件夹</span>
                    </button>
                </div>
            </div>
        `;
        
        document.body.appendChild(dialog);
        
        // 关闭对话框
        const closeDialog = () => {
            document.body.removeChild(dialog);
        };
        
        // 绑定事件
        dialog.querySelector('.dialog-close').addEventListener('click', closeDialog);
        dialog.querySelector('.dialog-overlay').addEventListener('click', closeDialog);
        
        // 文件选择
        dialog.querySelector('.file-button').addEventListener('click', async () => {
            closeDialog();
            await selectFile();
        });
        
        // 文件夹选择
        dialog.querySelector('.folder-button').addEventListener('click', async () => {
            closeDialog();
            await selectFolder();
        });
        
        // ESC键关闭
        const handleKeydown = (e) => {
            if (e.key === 'Escape') {
                closeDialog();
                document.removeEventListener('keydown', handleKeydown);
            }
        };
        document.addEventListener('keydown', handleKeydown);
    }
    
    // 选择文件
    async function selectFile() {
        if (!await initBackend()) {
            showError('后端未就绪，请稍后重试');
            return;
        }
        
        try {
            const selectedPath = await backend.SelectFile();
            if (selectedPath) {
                await handlePathSelection(selectedPath);
            }
        } catch (error) {
            console.error('选择文件失败:', error);
            showError('选择文件失败: ' + error);
        }
    }
    
    // 选择文件夹
    async function selectFolder() {
        if (!await initBackend()) {
            showError('后端未就绪，请稍后重试');
            return;
        }
        
        try {
            const selectedPath = await backend.SelectFolder();
            if (selectedPath) {
                await handlePathSelection(selectedPath);
            }
        } catch (error) {
            console.error('选择文件夹失败:', error);
            showError('选择文件夹失败: ' + error);
        }
    }
    
    // 显示错误信息
    function showError(message) {
        const errorDialog = document.createElement('div');
        errorDialog.className = 'error-dialog';
        errorDialog.innerHTML = `
            <div class="dialog-overlay"></div>
            <div class="dialog-content">
                <div class="dialog-header">
                    <h3>错误</h3>
                    <button class="dialog-close">&times;</button>
                </div>
                <div class="dialog-body">
                    <p>${message}</p>
                </div>
                <div class="dialog-footer">
                    <button class="dialog-button primary">确定</button>
                </div>
            </div>
        `;
        
        document.body.appendChild(errorDialog);
        
        const closeError = () => {
            document.body.removeChild(errorDialog);
        };
        
        errorDialog.querySelector('.dialog-close').addEventListener('click', closeError);
        errorDialog.querySelector('.dialog-overlay').addEventListener('click', closeError);
        errorDialog.querySelector('.dialog-button').addEventListener('click', closeError);
        
        // ESC键关闭
        const handleKeydown = (e) => {
            if (e.key === 'Escape') {
                closeError();
                document.removeEventListener('keydown', handleKeydown);
            }
        };
        document.addEventListener('keydown', handleKeydown);
    }
    
    // 点击选择文件 - 显示专业的选择对话框
    dropZone.addEventListener('click', () => {
        createFileSelectionDialog();
    });
    
    // 处理路径选择
    async function handlePathSelection(path) {
        if (!path) return;
        
        // 清空之前的文件列表
        fileList.innerHTML = '';
        
        // 获取文件/文件夹信息
        try {
            const stats = await backend.GetFileInfo(path);
            
            if (stats.error) {
                throw new Error(stats.error);
            }
            
            // 显示文件信息
            const fileItem = document.createElement('div');
            fileItem.className = 'file-item';
            fileItem.innerHTML = `
                <span class="file-icon">${stats.isDirectory ? '📁' : '📄'}</span>
                <span class="file-name">${stats.name}</span>
                <span class="file-size">${stats.sizeDisplay}</span>
            `;
            fileList.appendChild(fileItem);
            
            // 显示已选择文件区域
            selectedFiles.style.display = 'block';
            
            // 存储选择的完整路径
            window.currentSelectedPath = path;
            
            console.log('选择了路径:', path, '类型:', stats.isDirectory ? '文件夹' : '文件', '大小:', stats.sizeDisplay);
        } catch (error) {
            console.error('获取文件信息失败:', error);
            showError('获取文件信息失败: ' + error);
        }
    }
}

// 重置发送状态
window.resetSendState = async function() {
    if (!await initBackend()) {
        return;
    }
    
    try {
        // 清空选择的文件
        window.currentSelectedPath = null;
        document.getElementById('selectedFiles').style.display = 'none';
        document.getElementById('fileList').innerHTML = '';
        document.getElementById('sendStatus').textContent = '就绪';
        resetProgressBars();
    } catch (error) {
        console.error('重置发送状态失败:', error);
    }
}

// 初始化
document.addEventListener('DOMContentLoaded', async () => {
    console.log('前端初始化完成');
    
    // 设置文件选择功能
    setupFileSelection();
    
    // 等待后端绑定可用
    const checkBackend = setInterval(async () => {
        if (window.go && window.go.main && window.go.main.App) {
            clearInterval(checkBackend);
            backend = window.go.main.App;
            console.log('后端绑定已就绪');
        }
    }, 100);
    
    // 禁用缩放
    document.addEventListener('wheel', (e) => {
        if (e.ctrlKey) {
            e.preventDefault();
        }
    }, { passive: false });
    
    // 禁用键盘缩放
    document.addEventListener('keydown', (e) => {
        if (e.ctrlKey && (e.key === '+' || e.key === '-' || e.key === '0' || e.key === '=')) {
            e.preventDefault();
        }
    });
});
