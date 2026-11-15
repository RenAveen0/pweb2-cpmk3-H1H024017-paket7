// Konfirmasi hapus memakai Bootstrap Modal.
// Setiap tombol .btn-delete punya data-action (URL), data-name (nama data),
// dan data-note (catatan tambahan, opsional).
document.addEventListener('DOMContentLoaded', () => {
    const modalEl = document.getElementById('deleteModal');
    if (!modalEl) return;

    const modal = new bootstrap.Modal(modalEl);
    const form = document.getElementById('deleteForm');
    const nameEl = document.getElementById('deleteName');
    const noteEl = document.getElementById('deleteNote');

    document.querySelectorAll('.btn-delete').forEach((btn) => {
        btn.addEventListener('click', () => {
            form.action = btn.dataset.action;
            nameEl.textContent = btn.dataset.name;
            noteEl.textContent = btn.dataset.note || '';
            modal.show();
        });
    });
});
