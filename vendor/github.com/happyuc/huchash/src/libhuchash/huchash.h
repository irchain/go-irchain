/*
  This file is part of huchash.

  huchash is free software: you can redistribute it and/or modify
  it under the terms of the GNU General Public License as published by
  the Free Software Foundation, either version 3 of the License, or
  (at your option) any later version.

  huchash is distributed in the hope that it will be useful,
  but WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
  GNU General Public License for more details.

  You should have received a copy of the GNU General Public License
  along with huchash.  If not, see <http://www.gnu.org/licenses/>.
*/

/** @file huchash.h
* @date 2015
*/
#pragma once

#include <stdint.h>
#include <stdbool.h>
#include <string.h>
#include <stddef.h>
#include "compiler.h"

#define HUCHASH_REVISION 23
#define HUCHASH_DATASET_BYTES_INIT 1073741824U // 2**30
#define HUCHASH_DATASET_BYTES_GROWTH 8388608U  // 2**23
#define HUCHASH_CACHE_BYTES_INIT 1073741824U // 2**24
#define HUCHASH_CACHE_BYTES_GROWTH 131072U  // 2**17
#define HUCHASH_EPOCH_LENGTH 30000U
#define HUCHASH_MIX_BYTES 128
#define HUCHASH_HASH_BYTES 64
#define HUCHASH_DATASET_PARENTS 256
#define HUCHASH_CACHE_ROUNDS 3
#define HUCHASH_ACCESSES 64
#define HUCHASH_DAG_MAGIC_NUM_SIZE 8
#define HUCHASH_DAG_MAGIC_NUM 0xFEE1DEADBADDCAFE

#ifdef __cplusplus
extern "C" {
#endif

/// Type of a seedhash/blockhash e.t.c.
typedef struct huchash_h256 { uint8_t b[32]; } huchash_h256_t;

// convenience macro to statically initialize an h256_t
// usage:
// huchash_h256_t a = huchash_h256_static_init(1, 2, 3, ... )
// have to provide all 32 values. If you don't provide all the rest
// will simply be unitialized (not guranteed to be 0)
#define huchash_h256_static_init(...)			\
	{ {__VA_ARGS__} }

struct huchash_light;
typedef struct huchash_light* huchash_light_t;
struct huchash_full;
typedef struct huchash_full* huchash_full_t;
typedef int(*huchash_callback_t)(unsigned);

typedef struct huchash_return_value {
	huchash_h256_t result;
	huchash_h256_t mix_hash;
	bool success;
} huchash_return_value_t;

/**
 * Allocate and initialize a new huchash_light handler
 *
 * @param block_number   The block number for which to create the handler
 * @return               Newly allocated huchash_light handler or NULL in case of
 *                       ERRNOMEM or invalid parameters used for @ref huchash_compute_cache_nodes()
 */
huchash_light_t huchash_light_new(uint64_t block_number);
/**
 * Frees a previously allocated huchash_light handler
 * @param light        The light handler to free
 */
void huchash_light_delete(huchash_light_t light);
/**
 * Calculate the light client data
 *
 * @param light          The light client handler
 * @param header_hash    The header hash to pack into the mix
 * @param nonce          The nonce to pack into the mix
 * @return               an object of huchash_return_value_t holding the return values
 */
huchash_return_value_t huchash_light_compute(
	huchash_light_t light,
	huchash_h256_t const header_hash,
	uint64_t nonce
);

/**
 * Allocate and initialize a new huchash_full handler
 *
 * @param light         The light handler containing the cache.
 * @param callback      A callback function with signature of @ref huchash_callback_t
 *                      It accepts an unsigned with which a progress of DAG calculation
 *                      can be displayed. If all goes well the callback should return 0.
 *                      If a non-zero value is returned then DAG generation will stop.
 *                      Be advised. A progress value of 100 means that DAG creation is
 *                      almost complete and that this function will soon return succesfully.
 *                      It does not mean that the function has already had a succesfull return.
 * @return              Newly allocated huchash_full handler or NULL in case of
 *                      ERRNOMEM or invalid parameters used for @ref huchash_compute_full_data()
 */
huchash_full_t huchash_full_new(huchash_light_t light, huchash_callback_t callback);

/**
 * Frees a previously allocated huchash_full handler
 * @param full    The light handler to free
 */
void huchash_full_delete(huchash_full_t full);
/**
 * Calculate the full client data
 *
 * @param full           The full client handler
 * @param header_hash    The header hash to pack into the mix
 * @param nonce          The nonce to pack into the mix
 * @return               An object of huchash_return_value to hold the return value
 */
huchash_return_value_t huchash_full_compute(
	huchash_full_t full,
	huchash_h256_t const header_hash,
	uint64_t nonce
);
/**
 * Get a pointer to the full DAG data
 */
void const* huchash_full_dag(huchash_full_t full);
/**
 * Get the size of the DAG data
 */
uint64_t huchash_full_dag_size(huchash_full_t full);

/**
 * Calculate the seedhash for a given block number
 */
huchash_h256_t huchash_get_seedhash(uint64_t block_number);

#ifdef __cplusplus
}
#endif
