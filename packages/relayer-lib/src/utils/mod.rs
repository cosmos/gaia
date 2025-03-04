//! This module contains the utilities for relayer implementations.

use futures_timer::Delay;
use std::future::Future;
use std::time::{Duration, Instant};

/// Retries an operation until the condition is met or a timeout occurs.
pub async fn wait_for_condition<F, Fut>(
    timeout: Duration,
    interval: Duration,
    mut condition: F,
) -> anyhow::Result<()>
where
    F: FnMut() -> Fut + Send,
    Fut: Future<Output = anyhow::Result<bool>> + Send,
{
    let start = Instant::now();
    while start.elapsed() < timeout {
        if condition().await? {
            return Ok(());
        }

        tracing::debug!(
            "Condition not met. Waiting for {} seconds before retrying",
            interval.as_secs()
        );
        Delay::new(interval).await;
    }
    anyhow::bail!("Timeout exceeded")
}

pub mod cosmos;
pub mod eth_eureka;
