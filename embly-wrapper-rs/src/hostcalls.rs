use crate::context::EmblyCtx;
use crate::error::Result;
use log::debug;
use lucet_runtime::lucet_hostcalls;
use lucet_runtime::vmctx::Vmctx;
use lucet_wasi;
use lucet_wasi::{host, memory, wasm32};
use std::ffi::OsStr;
use std::os::unix::prelude::OsStrExt;
use std::time;

fn _read(
    vmctx: &mut Vmctx,
    id: wasm32::uintptr_t,
    payload: wasm32::uintptr_t,
    payload_len: wasm32::uintptr_t,
    ln: wasm32::uintptr_t,
) -> Result<()> {
    let mut ctx = vmctx.get_embed_ctx_mut::<EmblyCtx>();
    let bytes = memory::dec_slice_of_mut::<u8>(vmctx, payload, payload_len)?;
    let read = ctx.read(id as i32, bytes)?;
    memory::enc_usize_byref(vmctx, ln, read)?;
    Ok(())
}

fn _write(
    vmctx: &mut Vmctx,
    id: wasm32::uintptr_t,
    payload: wasm32::uintptr_t,
    payload_len: wasm32::uintptr_t,
    ln: wasm32::uintptr_t,
) -> Result<()> {
    let mut ctx = vmctx.get_embed_ctx_mut::<EmblyCtx>();
    let bytes = memory::dec_slice_of::<u8>(vmctx, payload, payload_len)?;
    let written = ctx.write(id as i32, bytes)?;
    memory::enc_usize_byref(vmctx, ln, written)?;
    debug!(
        "__write {:?}",
        (ctx.address, ctx.address_count, id, payload, payload_len, ln)
    );
    Ok(())
}

fn _spawn(
    vmctx: &mut Vmctx,
    name: wasm32::uintptr_t,
    name_len: wasm32::uintptr_t,
    id: wasm32::uintptr_t,
) -> Result<()> {
    let mut ctx = vmctx.get_embed_ctx_mut::<EmblyCtx>();
    let name = OsStr::from_bytes(memory::dec_slice_of::<u8>(vmctx, name, name_len)?);
    debug!("__spawn call {:?}", (name));
    let addr = ctx.spawn(name.to_str().unwrap())?;
    memory::enc_usize_byref(vmctx, id, addr as usize)?;
    Ok(())
}

fn _events(
    vmctx: &mut Vmctx,
    non_blocking: wasm32::uint8_t,
    timeout_s: wasm32::uint64_t,
    timeout_ns: wasm32::uint32_t,
    ids: wasm32::uintptr_t,
    ids_len: wasm32::uint32_t,
    ln: wasm32::uintptr_t,
) -> Result<()> {
    let mut ctx = vmctx.get_embed_ctx_mut::<EmblyCtx>();
    let timeout = if non_blocking == 0 {
        Some(time::Duration::new(timeout_s, timeout_ns))
    } else {
        None
    };
    let in_len = ids_len as usize;
    debug!("__events call {:?}", (in_len, timeout));
    let mut events = ctx.events_limited(timeout, in_len)?;
    debug!("__events got events {:?}", events);
    events.resize(in_len, 0);
    memory::enc_usize_byref(vmctx, ln, events.len())?;
    memory::enc_slice_of(vmctx, &events, ids)?;
    Ok(())
}

fn result_to_wasi_err(rest: Result<()>) -> u16 {
    (match rest {
        Ok(_) => host::__WASI_ESUCCESS,
        Err(err) => err.to_wasi_err(),
    } as u16)
}

lucet_hostcalls! {
    #[no_mangle] pub unsafe extern "C"
    fn __read(
        &mut vmctx,
        id: wasm32::uintptr_t,
        payload: wasm32::uintptr_t,
        payload_len: wasm32::uintptr_t,
        ln: wasm32::uintptr_t,
    ) -> wasm32::__wasi_errno_t {
        result_to_wasi_err(_read(vmctx, id, payload, payload_len, ln))
    }

    #[no_mangle] pub unsafe extern "C"
    fn __write(
        &mut vmctx,
        id: wasm32::uintptr_t,
        payload: wasm32::uintptr_t,
        payload_len: wasm32::uintptr_t,
        ln: wasm32::uintptr_t,
    ) -> wasm32::__wasi_errno_t {
        result_to_wasi_err(_write(vmctx, id, payload, payload_len, ln))
    }

    #[no_mangle] pub unsafe extern "C"
    fn __spawn(
        &mut vmctx,
        name: wasm32::uintptr_t,
        name_len: wasm32::uintptr_t,
        id: wasm32::uintptr_t,
    ) -> wasm32::__wasi_errno_t {
        result_to_wasi_err(_spawn(vmctx, name, name_len, id))
    }

    #[no_mangle] pub unsafe extern "C"
    fn __events(
        &mut vmctx,
        non_blocking: wasm32::uint8_t,
        timeout_s: wasm32::uint64_t,
        timeout_ns: wasm32::uint32_t,
        ids: wasm32::uintptr_t,
        ids_len: wasm32::uint32_t,
        ln: wasm32::uintptr_t,
    ) -> wasm32::__wasi_errno_t {
        result_to_wasi_err(_events(vmctx, non_blocking, timeout_s, timeout_ns, ids, ids_len, ln))
    }

}
