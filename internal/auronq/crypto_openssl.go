//go:build !windows && legacycrypto

package auronq

/*
#cgo LDFLAGS: -lcrypto
#include <openssl/evp.h>
#include <openssl/core_names.h>
#include <openssl/params.h>
#include <openssl/crypto.h>

EVP_PKEY* aq_mldsa87_generate() { return EVP_PKEY_Q_keygen(NULL, NULL, "ML-DSA-87"); }
int aq_get_seed(EVP_PKEY *p, unsigned char *out, size_t n, size_t *olen) { return EVP_PKEY_get_octet_string_param(p, OSSL_PKEY_PARAM_ML_DSA_SEED, out, n, olen); }
int aq_get_pub(EVP_PKEY *p, unsigned char *out, size_t n, size_t *olen) { return EVP_PKEY_get_octet_string_param(p, OSSL_PKEY_PARAM_PUB_KEY, out, n, olen); }
EVP_PKEY* aq_key_from_seed(unsigned char *seed, size_t n) {
  EVP_PKEY_CTX *ctx = EVP_PKEY_CTX_new_from_name(NULL, "ML-DSA-87", NULL); EVP_PKEY *p = NULL; if (!ctx) return NULL;
  if (EVP_PKEY_fromdata_init(ctx) <= 0) { EVP_PKEY_CTX_free(ctx); return NULL; }
  OSSL_PARAM params[2]; params[0] = OSSL_PARAM_construct_octet_string(OSSL_PKEY_PARAM_ML_DSA_SEED, seed, n); params[1] = OSSL_PARAM_construct_end();
  if (EVP_PKEY_fromdata(ctx, &p, EVP_PKEY_KEYPAIR, params) <= 0) p = NULL; EVP_PKEY_CTX_free(ctx); return p;
}
EVP_PKEY* aq_key_from_pub(unsigned char *pub, size_t n) {
  EVP_PKEY_CTX *ctx = EVP_PKEY_CTX_new_from_name(NULL, "ML-DSA-87", NULL); EVP_PKEY *p = NULL; if (!ctx) return NULL;
  if (EVP_PKEY_fromdata_init(ctx) <= 0) { EVP_PKEY_CTX_free(ctx); return NULL; }
  OSSL_PARAM params[2]; params[0] = OSSL_PARAM_construct_octet_string(OSSL_PKEY_PARAM_PUB_KEY, pub, n); params[1] = OSSL_PARAM_construct_end();
  if (EVP_PKEY_fromdata(ctx, &p, EVP_PKEY_PUBLIC_KEY, params) <= 0) p = NULL; EVP_PKEY_CTX_free(ctx); return p;
}
int aq_sign(EVP_PKEY *pkey, unsigned char *msg, size_t mlen, unsigned char **out, size_t *olen) {
  EVP_PKEY_CTX *ctx = EVP_PKEY_CTX_new_from_pkey(NULL,pkey,NULL); EVP_SIGNATURE *alg = EVP_SIGNATURE_fetch(NULL,"ML-DSA-87",NULL);
  if (!ctx || !alg) { EVP_PKEY_CTX_free(ctx); EVP_SIGNATURE_free(alg); return 0; }
  if (EVP_PKEY_sign_message_init(ctx,alg,NULL)<=0) { EVP_PKEY_CTX_free(ctx); EVP_SIGNATURE_free(alg); return 0; }
  size_t n=0; if(EVP_PKEY_sign(ctx,NULL,&n,msg,mlen)<=0){EVP_PKEY_CTX_free(ctx);EVP_SIGNATURE_free(alg);return 0;}
  unsigned char *buf=OPENSSL_malloc(n); if(!buf){EVP_PKEY_CTX_free(ctx);EVP_SIGNATURE_free(alg);return 0;}
  if(EVP_PKEY_sign(ctx,buf,&n,msg,mlen)<=0){OPENSSL_free(buf);EVP_PKEY_CTX_free(ctx);EVP_SIGNATURE_free(alg);return 0;}
  *out=buf;*olen=n;EVP_PKEY_CTX_free(ctx);EVP_SIGNATURE_free(alg);return 1;
}
int aq_verify(EVP_PKEY *pkey,unsigned char *msg,size_t mlen,unsigned char *sig,size_t slen){
  EVP_PKEY_CTX *ctx=EVP_PKEY_CTX_new_from_pkey(NULL,pkey,NULL); EVP_SIGNATURE *alg=EVP_SIGNATURE_fetch(NULL,"ML-DSA-87",NULL);
  if(!ctx||!alg){EVP_PKEY_CTX_free(ctx);EVP_SIGNATURE_free(alg);return 0;} if(EVP_PKEY_verify_message_init(ctx,alg,NULL)<=0){EVP_PKEY_CTX_free(ctx);EVP_SIGNATURE_free(alg);return 0;}
  int ok=EVP_PKEY_verify(ctx,sig,slen,msg,mlen);EVP_PKEY_CTX_free(ctx);EVP_SIGNATURE_free(alg);return ok==1;
}
void aq_free_pkey(EVP_PKEY *p){EVP_PKEY_free(p);} void aq_free_buf(unsigned char *p){OPENSSL_free(p);}
*/
import "C"

import (
	"errors"
	"fmt"
	"unsafe"
)

func GenerateMLDSA87() (seed, pub []byte, err error) {
	p := C.aq_mldsa87_generate()
	if p == nil {
		return nil, nil, errors.New("OpenSSL ML-DSA-87 key generation failed")
	}
	defer C.aq_free_pkey(p)
	seed = make([]byte, MLDSASeedSize)
	pub = make([]byte, MLDSA87PublicKeySize)
	var sn, pn C.size_t
	if C.aq_get_seed(p, (*C.uchar)(unsafe.Pointer(&seed[0])), C.size_t(len(seed)), &sn) != 1 || int(sn) != len(seed) {
		return nil, nil, errors.New("failed to export ML-DSA seed")
	}
	if C.aq_get_pub(p, (*C.uchar)(unsafe.Pointer(&pub[0])), C.size_t(len(pub)), &pn) != 1 || int(pn) != len(pub) {
		return nil, nil, errors.New("failed to export ML-DSA public key")
	}
	return seed, pub, nil
}
func PublicFromSeed(seed []byte) ([]byte, error) {
	if len(seed) != MLDSASeedSize {
		return nil, fmt.Errorf("invalid seed length: %d", len(seed))
	}
	p := C.aq_key_from_seed((*C.uchar)(unsafe.Pointer(&seed[0])), C.size_t(len(seed)))
	if p == nil {
		return nil, errors.New("failed to derive ML-DSA key from seed")
	}
	defer C.aq_free_pkey(p)
	pub := make([]byte, MLDSA87PublicKeySize)
	var pn C.size_t
	if C.aq_get_pub(p, (*C.uchar)(unsafe.Pointer(&pub[0])), C.size_t(len(pub)), &pn) != 1 || int(pn) != len(pub) {
		return nil, errors.New("failed to export derived public key")
	}
	return pub, nil
}
func SignMLDSA87(seed, msg []byte) ([]byte, error) {
	if len(seed) != MLDSASeedSize {
		return nil, errors.New("invalid ML-DSA seed")
	}
	if len(msg) == 0 {
		return nil, errors.New("refusing to sign empty message")
	}
	p := C.aq_key_from_seed((*C.uchar)(unsafe.Pointer(&seed[0])), C.size_t(len(seed)))
	if p == nil {
		return nil, errors.New("failed to load ML-DSA private key")
	}
	defer C.aq_free_pkey(p)
	var sig *C.uchar
	var slen C.size_t
	if C.aq_sign(p, (*C.uchar)(unsafe.Pointer(&msg[0])), C.size_t(len(msg)), &sig, &slen) != 1 {
		return nil, errors.New("ML-DSA signing failed")
	}
	defer C.aq_free_buf(sig)
	return C.GoBytes(unsafe.Pointer(sig), C.int(slen)), nil
}
func VerifyMLDSA87(pub, msg, sig []byte) bool {
	if len(pub) != MLDSA87PublicKeySize || len(sig) != MLDSA87SignatureSize || len(msg) == 0 {
		return false
	}
	p := C.aq_key_from_pub((*C.uchar)(unsafe.Pointer(&pub[0])), C.size_t(len(pub)))
	if p == nil {
		return false
	}
	defer C.aq_free_pkey(p)
	return C.aq_verify(p, (*C.uchar)(unsafe.Pointer(&msg[0])), C.size_t(len(msg)), (*C.uchar)(unsafe.Pointer(&sig[0])), C.size_t(len(sig))) == 1
}
