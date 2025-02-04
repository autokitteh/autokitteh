// Switch between 2 available modes: Default App, and Private App.

function toggleTab(id) {
    // Update the toggle buttons.
    const buttons = document.getElementsByClassName("toggle");
    for (let i = 0; i < buttons.length; i++) {
      buttons[i].classList.remove("active");
    }
    document.getElementById("toggle" + id).classList.add("active");
  
    // Update the tab contents.
    const tabs = document.getElementsByClassName("tab");
    for (let i = 0; i < tabs.length; i++) {
      tabs[i].classList.remove("active");
    }
    document.getElementById(id).classList.add("active");
  }
  
  window.onload = function () {
    const authSelect = document.getElementById('authType');
    if (authSelect) {
      const privateAppDiv = document.getElementById('privateApp');
      const clientIdField = document.getElementById('clientId');
      const clientSecretField = document.getElementById('clientSecret');

      authSelect.addEventListener('change', function () {
        if (this.value === 'oauthPrivate') {
          privateAppDiv.classList.remove('hidden');
          clientIdField.disabled = false;
          clientSecretField.disabled = false;
        } else {
          privateAppDiv.classList.add('hidden');
          clientIdField.disabled = true;
          clientSecretField.disabled = true;
        }
      });
      // Set the initial state according to the current selection
      authSelect.dispatchEvent(new Event('change'));
    } else {
      // Fallback to previous tab functionality if authType is not present
      toggleTab('tab1');
    }
  };
  