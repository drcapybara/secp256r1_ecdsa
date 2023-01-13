package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"math/big"
)

func run() {
	for i := 0; i < 100; i++ {
		rnd := rand.Reader
		data := make([]byte, 5242880) //5mb random data
		rnd.Read(data)
		key, sig, e := sign_message_schnorr(&data)
		println(verify_sig(&key, sig, e, &data))
	}
}

func sign_message_schnorr(msg *[]byte) (ecdsa.PublicKey, *big.Int, *big.Int) {
	secp256r1 := elliptic.P256() // aka secp256r1
	n := secp256r1.Params().Params().N

	g := ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     secp256r1.Params().Gx,
		Y:     secp256r1.Params().Gy,
	}

	// the secret key generated by the user
	rnd := rand.Reader
	x_bytes := make([]byte, 32)
	rnd.Read(x_bytes)
	pub_x, pub_y := g.ScalarBaseMult(x_bytes)
	y := ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     pub_x,
		Y:     pub_y,
	}
	x := big.NewInt(0).SetBytes(x_bytes)

	// random k from allowed set
	k_read := rand.Reader
	k_bytes := make([]byte, 32)
	k_read.Read(k_bytes)
	k := big.NewInt(0).SetBytes(k_bytes)
	k = k.Mod(k, n)

	r_x, _ := g.ScalarBaseMult(k.Bytes())
	e_hash := sha256.Sum256(append(r_x.Bytes(), *msg...))

	e := big.NewInt(0).SetBytes(e_hash[:32])
	xe := big.NewInt(0).Mul(x, e)

	s := k.Sub(k, xe)
	s = s.Mod(s, n)
	return y, s, e
}

/*
let r_v = g^s * y^e
let e_v Hash(r_v || M)

return true iff e_v = e
*/
func verify_sig(y *ecdsa.PublicKey, s, e *big.Int, msg *[]byte) bool {
	curve := elliptic.P256() // aka secp256r1

	g := ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     curve.Params().Gx,
		Y:     curve.Params().Gy,
	}

	gs_x, gs_y := g.ScalarBaseMult(s.Bytes())
	gy_x, gy_y := y.ScalarMult(y.X, y.Y, e.Bytes())

	r_x, _ := g.Add(gs_x, gs_y, gy_x, gy_y)

	e_v := sha256.Sum256(append(r_x.Bytes(), *msg...))
	return Equal(e_v[:32], e.Bytes())
}

// Compare byte arrays for equality
func Equal(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
