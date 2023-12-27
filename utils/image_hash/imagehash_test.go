package image_hash

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"io"
	"os"
	"testing"

	"github.com/nfnt/resize"
)

func TestBlurHash(t *testing.T) {

	f, err := os.Open("./data/images/9c5c925599424182ba50102b4c53ee2d.jpg")
	if err != nil {
		t.Fatal(err)
	}
	img, err := jpeg.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	resized := resize.Resize(64, 64, img, resize.Bilinear)
	bh, err := EncodeBlurHashFast(resized)
	if err != nil {
		t.Fatal(err)
	}
	if bh != "UM6KU]e9MyY5ysiwR*b^Qmi_j=i_IAV@nOsp" {
		t.Error(bh)
	}
}

func TestImageHash(t *testing.T) {
	hashTests := []struct {
		filename string
		phash64  string
		phash256 string
	}{
		{"../images-test/1.jpg", "p:a19d4eb2592613ed", "p:a1d39c6e0e39b2c759b626c90314ed78998372fe998c76e3a51456cd2872e78c"},
		{"../images-test/2.jpg", "p:a1dc4eb2592613ed", "p:a1d7986e0e39b2c759b626c90314ed78998372f7998c66e3a51456cda872e78c"},
		{"../images-test/3.jpg", "p:94c773b866d9a642", "p:9535c7427babbd1566eaf999aee47e5ab9a8ea66b5816b299604694a9602a948"},
		{"../images-test/4.jpg", "p:a19d4eb2592613ed", "p:a1d3986e0e39b2c759b626c90315ed78998372ff998c76e3a41456cd2872e78c"},
		{"../images-test/5.jpg", "p:a19d4eb2592613ed", "p:a1d3986e0e39b2c759b626c90315ed78998372ff998c76e3a41456cd2872e78c"},
		{"../images-test/6.jpg", "p:a19d4eb2592613ed", "p:a1d3986e0e39b2c759b626c90315ed78998372ff998c76e3a41456cd2872e78c"},
		{"../images-test/7.jpg", "p:a19d4eb2592613ed", "p:a1d3986e0e39b2c759b626c90315ed78998372ff998c76e3a41456cd2872e78c"},
		{"../images-test/8.jpg", "p:a19d4eb2592613ed", "p:a1d3986e0e39b2c759b626c90315ed78998372ff998c76e3a41456cd2872e78c"},
		{"../images-test/9.jpg", "p:a19d4eb2592613ed", "p:a1d3986e0e39b2c759b626c90315ed78998372ff998c76e3a41456cd2872e78c"},
		{"../images-test/10.jpg", "p:a19d4eb2592613ed", "p:a1d3986e0e39b2c759b626c90315ed78998372ff998c76e3a41456cd2872e78c"},
	}
	for _, h := range hashTests {
		fmt.Println()
		f, err := os.Open(h.filename)
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err = f.Close(); err != nil {
				t.Error(err)
			}
		}()

		img, err := jpeg.Decode(f)
		if err != nil {
			t.Fatal(err)
		}
		resized := resize.Resize(256, 256, img, resize.Bilinear)
		p256Alt, err := NewPHash256(resized)
		if err != nil {
			t.Fatal(err)
		}

		resized = resize.Resize(64, 64, img, resize.Bilinear)
		p64Alt, err := NewPHash64(resized)
		if err != nil {
			t.Fatal(err)
		}

		fmt.Println(p64Alt.String(), p256Alt.String())

		//   1  p:a19d4eb2592613ed p:a1d39c6e0e39b2c759b626c90314ed78998372fe998c76e3a51456cd2872e78c   1
		//   2  p:a1dc4eb2592613ed p:a1d7986e0e39b2c759b626c90314ed78998372f7998c66e3a51456cda872e78c   2
		//   3  p:94c773b866d9a642 p:9535c7427babbd1566eaf999aee47e5ab9a8ea66b5816b299604694a9602a948   3
		//   4  p:a19d4eb2592613ed p:a1d3986e0e39b2c759b626c90315ed78998372ff998c76e3a41456cd2872e78c   4
		//   5  p:948773b866d9a652 p:9535c7427baabd1566eaf999aee47e5ab9a8ea66b5816a399604695a9602a948   5
		//   6  p:a19d4eb2592613ed p:a1d39d6e0e39b2c759b626c90314ed7c998372f6999c6663a51456cd2872e78c   6
		//   7  p:a19b0eb2d936836d p:a1d38a6c0e39b2c3c9be364b83156dfe8b0970b69930464fa43557c9b876678b   7
		//   8  p:948773b866d9a652 p:9535c7427baabd1566eaf999aee47e5ab9a8ea66b5816a399604695a9602a948   8
		//   9  p:94c77ab8a6d9a648 p:9c3dc7527baabd15e6ead9b8ae655a5ab5a1ea6635896a359604694b9600294b   9
		//  10  p:948773b866d9a652 p:9535c7427baabd1566eaf999aee47e5ab9a8ea66b5816a399604695a9602a948   10

		//
		//p256Alt, err := NewPHash256Alt(resized)
		//if err != nil {
		//	t.Fatal(err)
		//}
		//p256, err := NewPHash256(resized)
		//if err != nil {
		//	t.Fatal(err)
		//}
		//
		//resized = resize.Resize(64, 64, img, resize.Bilinear)
		//p64Alt, err := NewPHash64Alt(resized)
		//if err != nil {
		//	t.Fatal(err)
		//}
		//p64, err := NewPHash64(resized)
		//if err != nil {
		//	t.Fatal(err)
		//}
		//if h.phash256 != p256Alt.String() && h.phash256 != p256.String() {
		//	t.Errorf("expected \t%s, got \t%s and \t%s", h.phash256, p256Alt, p256)
		//}
		//if h.phash64 != p64Alt.String() && h.phash64 != p64.String() {
		//	t.Errorf("expected \t%s, got \t%s and \t%s", h.phash64, p64Alt, p64)
		//}
	}

}

func BenchmarkPHash64(b *testing.B) {
	f, err := os.Open("../assets/a1.jpg")
	if err != nil {
		b.Fatal(err)
	}
	defer func() {
		if err = f.Close(); err != nil {
			b.Error(err)
		}
	}()
	buf, err := io.ReadAll(f)
	if err != nil {
		b.Fatal(err)
	}
	resized, err := jpeg.Decode(bytes.NewReader(buf))
	if err != nil {
		b.Fatal(err)
	}
	resized = resize.Resize(64, 64, resized, resize.Bicubic)
	b.ReportAllocs()
	b.ResetTimer()

	b.Run("Fast", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if _, err = NewPHash64(resized); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("FastAlt", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if _, err = NewPHash64(resized); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Fast-Parallel", func(b *testing.B) {
		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				if _, err = NewPHash64(resized); err != nil {
					b.Fatal(err)
				}
			}
		})
	})

}

func BenchmarkPHash256(b *testing.B) {
	f, err := os.Open("../assets/a1.jpg")
	if err != nil {
		b.Fatal(err)
	}
	defer func() {
		if err = f.Close(); err != nil {
			b.Error(err)
		}
	}()
	buf, err := io.ReadAll(f)
	if err != nil {
		b.Fatal(err)
	}
	resized, err := jpeg.Decode(bytes.NewReader(buf))
	if err != nil {
		b.Fatal(err)
	}
	resized = resize.Resize(256, 256, resized, resize.Bicubic)
	b.ReportAllocs()
	b.ResetTimer()

	b.Run("Fast", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if _, err = NewPHash256(resized); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("FastAlt", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if _, err = NewPHash256(resized); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Fast-Parallel", func(b *testing.B) {
		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				if _, err = NewPHash256(resized); err != nil {
					b.Fatal(err)
				}
			}
		})
	})

}

func BenchmarkBlurHash100(b *testing.B) {
	f, err := os.Open("../assets/a1.jpg")
	if err != nil {
		b.Fatal(err)
	}
	img, err := jpeg.Decode(f)
	if err != nil {
		b.Fatal(err)
	}

	resized := resize.Resize(64, 64, img, resize.Bilinear)
	b.Run("BlurHash", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			bh, err := EncodeBlurHashFast(resized)
			if err != nil {
				b.Fatal(err)
			}
			_ = bh
		}
	})
	//b.Run("BlurHashFast", func(b *testing.B) {
	//	for i := 0; i < b.N; i++ {
	//		bh, err := EncodeBlurHashFast(resized)
	//		if err != nil {
	//			b.Fatal(err)
	//		}
	//		_ = bh
	//	}
	//})

}
