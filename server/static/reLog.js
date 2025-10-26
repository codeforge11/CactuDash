let isRegisterMode = false;

async function checkAM() {
    try {
        const response = await fetch('/auth-mode');
        const data = await response.json();
        isRegisterMode = !data.dbExists;
        setUIMode();
    } catch (error) {
        console.error('Error checking auth mode:', error);
        showError('Failed to check authentication mode');
    }
}

function setUIMode() {
    const formTitle = document.getElementById('form-title');
    const submitText = document.getElementById('submit-text');
    const passwd2Container = document.getElementById('passwd2-container');
    const warning = document.getElementById('warning');
    const authForm = document.getElementById('auth-form');
    const usernameInput = document.getElementById('username')
    
    if (isRegisterMode) {
        formTitle.textContent = 'Register for "admin" account';
        submitText.textContent = 'Set Password & Restart';
        passwd2Container.classList.remove('hidden');
        warning.classList.add('show');
        authForm.action = '/register';
        usernameInput.value = 'admin';
        usernameInput.readOnly = true;
        usernameInput.style.display = 'none';
    } else {
        formTitle.textContent = 'Login';
        submitText.textContent = 'Login';
        passwd2Container.classList.add('hidden');
        warning.classList.remove('show');
        authForm.action = '/auth';
        usernameInput.readOnly = false;
    }
}

function showError(message) {
    const errorLabel = document.getElementById('errorMessage');

    errorLabel.textContent = message;
    errorLabel.style.display = 'block';
    
    setTimeout(() => errorLabel.style.display = 'none', 3000);
}

document.addEventListener('DOMContentLoaded', () => {
    checkAM();
    
    document.getElementById('auth-form').addEventListener('submit', (e) => {
        if (!isRegisterMode) return;
        
        const passwd1 = document.getElementById('password').value;
        const passwd2 = document.getElementById('passwd2').value;
        
        if (passwd1.length < 8) {
            e.preventDefault();
            showError('Password must be at least 8 characters!');
            return;
        }
        
        if (passwd1 !== passwd2) {
            e.preventDefault();
            showError('Passwords must be the same!');
        }
    });
    
    ['password', 'passwd2'].forEach(id => {
        document.getElementById(id).addEventListener('input', () => {
            document.getElementById('errorMessage').style.display = 'none';
        });
    });
});