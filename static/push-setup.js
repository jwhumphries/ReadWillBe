let swRegistration = null;

function urlBase64ToUint8Array(base64String) {
  const padding = '='.repeat((4 - base64String.length % 4) % 4);
  const base64 = (base64String + padding)
    .replace(/\-/g, '+')
    .replace(/_/g, '/');

  const rawData = window.atob(base64);
  const outputArray = new Uint8Array(rawData.length);

  for (let i = 0; i < rawData.length; ++i) {
    outputArray[i] = rawData.charCodeAt(i);
  }
  return outputArray;
}

async function registerServiceWorker() {
  if (!('serviceWorker' in navigator)) {
    console.log('Service Worker not supported');
    return null;
  }

  try {
    const registration = await navigator.serviceWorker.register('/serviceWorker.js', {
      scope: '/'
    });
    console.log('Service Worker registered');
    return registration;
  } catch (error) {
    console.error('Service Worker registration failed:', error);
    return null;
  }
}

async function checkSubscriptionStatus() {
  if (!swRegistration) {
    return;
  }

  const subscription = await swRegistration.pushManager.getSubscription();
  const enableBtn = document.getElementById('enable-push-btn');
  const disableBtn = document.getElementById('disable-push-btn');
  const badge = document.getElementById('subscription-badge');

  if (!enableBtn || !disableBtn || !badge) {
    return;
  }

  if (subscription) {
    badge.textContent = 'Subscribed';
    badge.classList.remove('badge-neutral');
    badge.classList.add('badge-success');
    enableBtn.classList.add('hidden');
    disableBtn.classList.remove('hidden');
  } else {
    badge.textContent = 'Not subscribed';
    badge.classList.remove('badge-neutral', 'badge-success');
    badge.classList.add('badge-ghost');
    enableBtn.classList.remove('hidden');
    disableBtn.classList.add('hidden');
  }
}

async function enablePushNotifications() {
  try {
    const permission = await Notification.requestPermission();
    if (permission !== 'granted') {
      alert('Notification permission denied');
      return;
    }

    const vapidKeyElement = document.querySelector('[data-vapid-key]');
    const vapidPublicKey = vapidKeyElement?.getAttribute('data-vapid-key');

    if (!vapidPublicKey) {
      alert('VAPID key not configured');
      return;
    }

    const subscription = await swRegistration.pushManager.subscribe({
      userVisibleOnly: true,
      applicationServerKey: urlBase64ToUint8Array(vapidPublicKey)
    });

    const response = await fetch('/push/subscribe', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        endpoint: subscription.endpoint,
        keys: {
          p256dh: btoa(String.fromCharCode.apply(null, new Uint8Array(subscription.getKey('p256dh')))),
          auth: btoa(String.fromCharCode.apply(null, new Uint8Array(subscription.getKey('auth'))))
        }
      })
    });

    if (response.ok) {
      checkSubscriptionStatus();
    } else {
      alert('Failed to save subscription');
    }
  } catch (error) {
    console.error('Error enabling push notifications:', error);
    alert('Failed to enable push notifications');
  }
}

async function disablePushNotifications() {
  try {
    const subscription = await swRegistration.pushManager.getSubscription();
    if (subscription) {
      await subscription.unsubscribe();

      await fetch('/push/unsubscribe', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          endpoint: subscription.endpoint
        })
      });
    }

    checkSubscriptionStatus();
  } catch (error) {
    console.error('Error disabling push notifications:', error);
    alert('Failed to disable push notifications');
  }
}

document.addEventListener('DOMContentLoaded', async () => {
  swRegistration = await registerServiceWorker();
  if (swRegistration) {
    await checkSubscriptionStatus();
  }
});
