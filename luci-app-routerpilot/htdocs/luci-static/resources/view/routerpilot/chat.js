'use strict';

var baseUrl = 'http://' + window.location.hostname + ':8080';

function addMessage(who, text, cls) {
	var box = document.getElementById('rp-chat-box');
	if (!box) return;
	var d = new Date();
	var ts = d.toLocaleTimeString();
	var entry = document.createElement('div');
	entry.style.marginBottom = '6px';
	entry.style.padding = '4px 8px';
	entry.style.borderRadius = '4px';

	if (who === 'me') {
		entry.style.background = '#e8f0fe';
		entry.innerHTML = '<strong>' + String(_('You')) + ' [' + ts + ']:</strong> ' + escapeHTML(text);
	} else if (who === 'bot') {
		entry.style.background = '#f0fdf4';
		entry.innerHTML = '<strong>' + String(_('Bot')) + ' [' + ts + ']:</strong> ' + (cls === 'error' ? '<span style="color:#d73a49">' : '') + escapeHTML(text) + (cls === 'error' ? '</span>' : '');
	} else {
		entry.style.background = '#f6f8fa';
		entry.style.color = '#656d76';
		entry.innerHTML = escapeHTML(text);
	}

	box.appendChild(entry);
	box.scrollTop = box.scrollHeight;
}

function appendResult(data) {
	var box = document.getElementById('rp-chat-box');
	if (!box) return;
	var entry = document.createElement('div');
	entry.style.marginBottom = '6px';
	entry.style.padding = '4px 8px';
	entry.style.borderRadius = '4px';
	entry.style.background = '#f0fdf4';

	var text;
	if (data.result) {
		text = JSON.stringify(data.result.output || data.result, null, 2);
	} else if (data.output) {
		text = JSON.stringify(data.output, null, 2);
	} else {
		text = JSON.stringify(data, null, 2);
	}

	entry.innerHTML = '<pre style="margin:0;font-size:12px;white-space:pre-wrap">' + escapeHTML(text) + '</pre>';
	box.appendChild(entry);
	box.scrollTop = box.scrollHeight;
}

function quickChat(text) {
	addMessage('me', text);

	var parts = text.split(/\s+/);
	var intentName = parts[0];
	var intentArgs = {};

	if (parts.length > 1) {
		intentArgs.target = parts.slice(1).join(' ');
	}

	addMessage('system', '→ executing ' + intentName + '...');

	fetch(baseUrl + '/intent', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ intent: intentName, args: intentArgs })
	}).then(function(r) {
		return r.json().then(function(data) { return { status: r.status, data: data }; });
	}).then(function(resp) {
		if (resp.status === 403 && resp.data.error === 'safety_confirmation_required') {
			addMessage('bot', '⚠ ' + String(_('This intent requires safety confirmation. Use CLI or set a higher risk level in Settings → General.')), 'error');
			return;
		}
		if (!resp.data || resp.data.state === 'failed') {
			addMessage('bot', '✗ ' + String(_('Failed:')) + ' ' + (resp.data.error || 'unknown error'), 'error');
			return;
		}
		appendResult(resp.data);
	}).catch(function(err) {
		addMessage('bot', '✗ ' + String(_('Connection error:')) + ' ' + err.message + '. Is the RouterPilot daemon running on port 8080?', 'error');
	});
}

function escapeHTML(s) {
	if (typeof s !== 'string') s = String(s);
	return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
}

return L.view.extend({
	render: function() {
		var self = this;
		var view = E('div', { 'class': 'cbi-map' }, [
			E('h2', { 'class': 'cbi-map-title' }, _('RouterPilot — Chat'))
		]);

		var chatBox = E('div', {
			'id': 'rp-chat-box',
			'style': 'background:#fff;border:1px solid #ccc;border-radius:4px;padding:8px;height:50vh;overflow-y:auto;margin-bottom:8px;font-family:monospace;font-size:13px;line-height:1.5;white-space:pre-wrap'
		});
		view.appendChild(chatBox);

		var wrapper = E('div', { 'id': 'rp-chat-ui' }, [
			E('div', { 'class': 'cbi-value', 'style': 'margin-bottom:0' }, [
				E('label', { 'class': 'cbi-value-title' }, _('Message')),
				E('div', { 'class': 'cbi-value-field', 'style': 'display:flex;gap:8px' }, [
					E('input', {
						'id': 'rp-chat-input',
						'type': 'text',
						'style': 'flex:1;max-width:none',
						'placeholder': _('Type a command or question...')
					}),
					E('button', {
						'id': 'rp-chat-send',
						'class': 'btn cbi-button cbi-button-action'
					}, _('Send'))
				])
			]),
			E('div', { 'class': 'cbi-value', 'style': 'margin-bottom:0' }, [
				E('label', { 'class': 'cbi-value-title' }, _('Quick')),
				E('div', { 'class': 'cbi-value-field', 'style': 'display:flex;gap:4px;flex-wrap:wrap' }, [
					E('button', { 'class': 'btn cbi-button cbi-button-neutral', 'data-quick': 'ping 8.8.8.8' }, 'ping 8.8.8.8'),
					E('button', { 'class': 'btn cbi-button cbi-button-neutral', 'data-quick': 'diagnose' }, 'diagnose'),
					E('button', { 'class': 'btn cbi-button cbi-button-neutral', 'data-quick': 'system.info' }, 'system.info'),
					E('button', { 'class': 'btn cbi-button cbi-button-neutral', 'data-quick': 'dns.lookup google.com' }, 'dns.lookup'),
					E('button', { 'class': 'btn cbi-button cbi-button-neutral', 'data-quick': 'wifi.scan' }, 'wifi.scan'),
					E('button', { 'class': 'btn cbi-button cbi-button-neutral', 'data-quick': 'suggest fixes' }, 'suggest fixes')
				])
			])
		]);
		view.appendChild(wrapper);

		// Delegate events
		setTimeout(function() { self.bindEvents(); }, 0);

		return view;
	},

	bindEvents: function() {
		var ui = document.getElementById('rp-chat-ui');
		if (!ui) return;

		ui.addEventListener('click', function(e) {
			var el = e.target;
			if (el.id === 'rp-chat-send') {
				sendChat();
			} else if (el.hasAttribute('data-quick')) {
				quickChat(el.getAttribute('data-quick'));
			}
		});

		var input = document.getElementById('rp-chat-input');
		if (input) {
			input.addEventListener('keydown', function(ev) {
				if (ev.key === 'Enter') sendChat();
			});
		}
	}
});

function sendChat() {
	var input = document.getElementById('rp-chat-input');
	if (!input) return;
	var text = input.value.trim();
	if (!text) return;
	input.value = '';
	quickChat(text);
}
