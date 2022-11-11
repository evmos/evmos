package keeper

import (
	"encoding/json"
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
)

func TestBuildContractAddress(t *testing.T) {
	x, y := sdk.GetConfig().GetBech32AccountAddrPrefix(), sdk.GetConfig().GetBech32AccountPubPrefix()
	t.Cleanup(func() {
		sdk.GetConfig().SetBech32PrefixForAccount(x, y)
	})
	sdk.GetConfig().SetBech32PrefixForAccount("purple", "purple")

	// test vectors generated via cosmjs: https://github.com/cosmos/cosmjs/pull/1253/files
	type Spec struct {
		In struct {
			Checksum tmbytes.HexBytes `json:"checksum"`
			Creator  sdk.AccAddress   `json:"creator"`
			Salt     tmbytes.HexBytes `json:"salt"`
			Msg      string           `json:"msg"`
		} `json:"in"`
		Out struct {
			Address sdk.AccAddress `json:"address"`
		} `json:"out"`
	}
	var specs []Spec
	require.NoError(t, json.Unmarshal([]byte(goldenMasterPredictableContractAddr), &specs))
	require.NotEmpty(t, specs)
	for i, spec := range specs {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			// when
			gotAddr := BuildContractAddressPredictable(spec.In.Checksum, spec.In.Creator, spec.In.Salt.Bytes(), []byte(spec.In.Msg))

			require.Equal(t, spec.Out.Address.String(), gotAddr.String())
			require.NoError(t, sdk.VerifyAddressFormat(gotAddr))
		})
	}
}

const goldenMasterPredictableContractAddr = `[
  {
    "in": {
      "checksum": "13a1fc994cc6d1c81b746ee0c0ff6f90043875e0bf1d9be6b7d779fc978dc2a5",
      "creator": "purple1nxvenxve42424242hwamhwamenxvenxvhxf2py",
      "creatorData": "9999999999aaaaaaaaaabbbbbbbbbbcccccccccc",
      "salt": "61",
      "msg": null
    },
    "intermediate": {
      "key": "7761736d00000000000000002013a1fc994cc6d1c81b746ee0c0ff6f90043875e0bf1d9be6b7d779fc978dc2a500000000000000149999999999aaaaaaaaaabbbbbbbbbbcccccccccc0000000000000001610000000000000000",
      "addressData": "5e865d3e45ad3e961f77fd77d46543417ced44d924dc3e079b5415ff6775f847"
    },
    "out": {
      "address": "purple1t6r960j945lfv8mhl4mage2rg97w63xeynwrupum2s2l7em4lprs9ce5hk"
    }
  },
  {
    "in": {
      "checksum": "13a1fc994cc6d1c81b746ee0c0ff6f90043875e0bf1d9be6b7d779fc978dc2a5",
      "creator": "purple1nxvenxve42424242hwamhwamenxvenxvhxf2py",
      "creatorData": "9999999999aaaaaaaaaabbbbbbbbbbcccccccccc",
      "salt": "61",
      "msg": "{}"
    },
    "intermediate": {
      "key": "7761736d00000000000000002013a1fc994cc6d1c81b746ee0c0ff6f90043875e0bf1d9be6b7d779fc978dc2a500000000000000149999999999aaaaaaaaaabbbbbbbbbbcccccccccc00000000000000016100000000000000027b7d",
      "addressData": "0995499608947a5281e2c7ebd71bdb26a1ad981946dad57f6c4d3ee35de77835"
    },
    "out": {
      "address": "purple1px25n9sgj3a99q0zcl4awx7my6s6mxqegmdd2lmvf5lwxh080q6suttktr"
    }
  },
  {
    "in": {
      "checksum": "13a1fc994cc6d1c81b746ee0c0ff6f90043875e0bf1d9be6b7d779fc978dc2a5",
      "creator": "purple1nxvenxve42424242hwamhwamenxvenxvhxf2py",
      "creatorData": "9999999999aaaaaaaaaabbbbbbbbbbcccccccccc",
      "salt": "61",
      "msg": "{\"some\":123,\"structure\":{\"nested\":[\"ok\",true]}}"
    },
    "intermediate": {
      "key": "7761736d00000000000000002013a1fc994cc6d1c81b746ee0c0ff6f90043875e0bf1d9be6b7d779fc978dc2a500000000000000149999999999aaaaaaaaaabbbbbbbbbbcccccccccc000000000000000161000000000000002f7b22736f6d65223a3132332c22737472756374757265223a7b226e6573746564223a5b226f6b222c747275655d7d7d",
      "addressData": "83326e554723b15bac664ceabc8a5887e27003abe9fbd992af8c7bcea4745167"
    },
    "out": {
      "address": "purple1svexu428ywc4htrxfn4tezjcsl38qqata8aany4033auafr529ns4v254c"
    }
  },
  {
    "in": {
      "checksum": "13a1fc994cc6d1c81b746ee0c0ff6f90043875e0bf1d9be6b7d779fc978dc2a5",
      "creator": "purple1nxvenxve42424242hwamhwamenxvenxvhxf2py",
      "creatorData": "9999999999aaaaaaaaaabbbbbbbbbbcccccccccc",
      "salt": "aabbccddeeffffeeddbbccddaa66551155aaaabbcc787878789900aabbccddeeffffeeddbbccddaa66551155aaaabbcc787878789900aabbbbcc221100acadae",
      "msg": null
    },
    "intermediate": {
      "key": "7761736d00000000000000002013a1fc994cc6d1c81b746ee0c0ff6f90043875e0bf1d9be6b7d779fc978dc2a500000000000000149999999999aaaaaaaaaabbbbbbbbbbcccccccccc0000000000000040aabbccddeeffffeeddbbccddaa66551155aaaabbcc787878789900aabbccddeeffffeeddbbccddaa66551155aaaabbcc787878789900aabbbbcc221100acadae0000000000000000",
      "addressData": "9384c6248c0bb171e306fd7da0993ec1e20eba006452a3a9e078883eb3594564"
    },
    "out": {
      "address": "purple1jwzvvfyvpwchrccxl476pxf7c83qawsqv3f2820q0zyrav6eg4jqdcq7gc"
    }
  },
  {
    "in": {
      "checksum": "13a1fc994cc6d1c81b746ee0c0ff6f90043875e0bf1d9be6b7d779fc978dc2a5",
      "creator": "purple1nxvenxve42424242hwamhwamenxvenxvhxf2py",
      "creatorData": "9999999999aaaaaaaaaabbbbbbbbbbcccccccccc",
      "salt": "aabbccddeeffffeeddbbccddaa66551155aaaabbcc787878789900aabbccddeeffffeeddbbccddaa66551155aaaabbcc787878789900aabbbbcc221100acadae",
      "msg": "{}"
    },
    "intermediate": {
      "key": "7761736d00000000000000002013a1fc994cc6d1c81b746ee0c0ff6f90043875e0bf1d9be6b7d779fc978dc2a500000000000000149999999999aaaaaaaaaabbbbbbbbbbcccccccccc0000000000000040aabbccddeeffffeeddbbccddaa66551155aaaabbcc787878789900aabbccddeeffffeeddbbccddaa66551155aaaabbcc787878789900aabbbbcc221100acadae00000000000000027b7d",
      "addressData": "9a8d5f98fb186825401a26206158e7a1213311a9b6a87944469913655af52ffb"
    },
    "out": {
      "address": "purple1n2x4lx8mrp5z2sq6ycsxzk885ysnxydfk658j3zxnyfk2kh49lasesxf6j"
    }
  },
  {
    "in": {
      "checksum": "13a1fc994cc6d1c81b746ee0c0ff6f90043875e0bf1d9be6b7d779fc978dc2a5",
      "creator": "purple1nxvenxve42424242hwamhwamenxvenxvhxf2py",
      "creatorData": "9999999999aaaaaaaaaabbbbbbbbbbcccccccccc",
      "salt": "aabbccddeeffffeeddbbccddaa66551155aaaabbcc787878789900aabbccddeeffffeeddbbccddaa66551155aaaabbcc787878789900aabbbbcc221100acadae",
      "msg": "{\"some\":123,\"structure\":{\"nested\":[\"ok\",true]}}"
    },
    "intermediate": {
      "key": "7761736d00000000000000002013a1fc994cc6d1c81b746ee0c0ff6f90043875e0bf1d9be6b7d779fc978dc2a500000000000000149999999999aaaaaaaaaabbbbbbbbbbcccccccccc0000000000000040aabbccddeeffffeeddbbccddaa66551155aaaabbcc787878789900aabbccddeeffffeeddbbccddaa66551155aaaabbcc787878789900aabbbbcc221100acadae000000000000002f7b22736f6d65223a3132332c22737472756374757265223a7b226e6573746564223a5b226f6b222c747275655d7d7d",
      "addressData": "932f07bc53f7d0b0b43cb5a54ac3e245b205e6ae6f7c1d991dc6af4a2ff9ac18"
    },
    "out": {
      "address": "purple1jvhs00zn7lgtpdpukkj54slzgkeqte4wda7pmxgac6h55tle4svq8cmp60"
    }
  },
  {
    "in": {
      "checksum": "13a1fc994cc6d1c81b746ee0c0ff6f90043875e0bf1d9be6b7d779fc978dc2a5",
      "creator": "purple1nxvenxve42424242hwamhwamenxvenxvmhwamhwaamhwamhwlllsatsy6m",
      "creatorData": "9999999999aaaaaaaaaabbbbbbbbbbccccccccccddddddddddeeeeeeeeeeffff",
      "salt": "61",
      "msg": null
    },
    "intermediate": {
      "key": "7761736d00000000000000002013a1fc994cc6d1c81b746ee0c0ff6f90043875e0bf1d9be6b7d779fc978dc2a500000000000000209999999999aaaaaaaaaabbbbbbbbbbccccccccccddddddddddeeeeeeeeeeffff0000000000000001610000000000000000",
      "addressData": "9725e94f528d8b78d33c25f3dfcd60e6142d8be60ab36f6a5b59036fd51560db"
    },
    "out": {
      "address": "purple1juj7jn6j3k9h35euyhealntquc2zmzlxp2ek76jmtypkl4g4vrdsfwmwxk"
    }
  },
  {
    "in": {
      "checksum": "13a1fc994cc6d1c81b746ee0c0ff6f90043875e0bf1d9be6b7d779fc978dc2a5",
      "creator": "purple1nxvenxve42424242hwamhwamenxvenxvmhwamhwaamhwamhwlllsatsy6m",
      "creatorData": "9999999999aaaaaaaaaabbbbbbbbbbccccccccccddddddddddeeeeeeeeeeffff",
      "salt": "61",
      "msg": "{}"
    },
    "intermediate": {
      "key": "7761736d00000000000000002013a1fc994cc6d1c81b746ee0c0ff6f90043875e0bf1d9be6b7d779fc978dc2a500000000000000209999999999aaaaaaaaaabbbbbbbbbbccccccccccddddddddddeeeeeeeeeeffff00000000000000016100000000000000027b7d",
      "addressData": "b056e539bbaf447ba18f3f13b792970111fc78933eb6700f4d227b5216d63658"
    },
    "out": {
      "address": "purple1kptw2wdm4az8hgv08ufm0y5hqyglc7yn86m8qr6dyfa4y9kkxevqmkm9q3"
    }
  },
  {
    "in": {
      "checksum": "13a1fc994cc6d1c81b746ee0c0ff6f90043875e0bf1d9be6b7d779fc978dc2a5",
      "creator": "purple1nxvenxve42424242hwamhwamenxvenxvmhwamhwaamhwamhwlllsatsy6m",
      "creatorData": "9999999999aaaaaaaaaabbbbbbbbbbccccccccccddddddddddeeeeeeeeeeffff",
      "salt": "61",
      "msg": "{\"some\":123,\"structure\":{\"nested\":[\"ok\",true]}}"
    },
    "intermediate": {
      "key": "7761736d00000000000000002013a1fc994cc6d1c81b746ee0c0ff6f90043875e0bf1d9be6b7d779fc978dc2a500000000000000209999999999aaaaaaaaaabbbbbbbbbbccccccccccddddddddddeeeeeeeeeeffff000000000000000161000000000000002f7b22736f6d65223a3132332c22737472756374757265223a7b226e6573746564223a5b226f6b222c747275655d7d7d",
      "addressData": "6c98434180f052294ff89fb6d2dae34f9f4468b0b8e6e7c002b2a44adee39abd"
    },
    "out": {
      "address": "purple1djvyxsvq7pfzjnlcn7md9khrf705g69shrnw0sqzk2jy4hhrn27sjh2ysy"
    }
  },
  {
    "in": {
      "checksum": "13a1fc994cc6d1c81b746ee0c0ff6f90043875e0bf1d9be6b7d779fc978dc2a5",
      "creator": "purple1nxvenxve42424242hwamhwamenxvenxvmhwamhwaamhwamhwlllsatsy6m",
      "creatorData": "9999999999aaaaaaaaaabbbbbbbbbbccccccccccddddddddddeeeeeeeeeeffff",
      "salt": "aabbccddeeffffeeddbbccddaa66551155aaaabbcc787878789900aabbccddeeffffeeddbbccddaa66551155aaaabbcc787878789900aabbbbcc221100acadae",
      "msg": null
    },
    "intermediate": {
      "key": "7761736d00000000000000002013a1fc994cc6d1c81b746ee0c0ff6f90043875e0bf1d9be6b7d779fc978dc2a500000000000000209999999999aaaaaaaaaabbbbbbbbbbccccccccccddddddddddeeeeeeeeeeffff0000000000000040aabbccddeeffffeeddbbccddaa66551155aaaabbcc787878789900aabbccddeeffffeeddbbccddaa66551155aaaabbcc787878789900aabbbbcc221100acadae0000000000000000",
      "addressData": "0aaf1c31c5d529d21d898775bc35b3416f47bfd99188c334c6c716102cbd3101"
    },
    "out": {
      "address": "purple1p2h3cvw9655ay8vfsa6mcddng9h5007ejxyvxdxxcutpqt9axyqsagmmay"
    }
  },
  {
    "in": {
      "checksum": "13a1fc994cc6d1c81b746ee0c0ff6f90043875e0bf1d9be6b7d779fc978dc2a5",
      "creator": "purple1nxvenxve42424242hwamhwamenxvenxvmhwamhwaamhwamhwlllsatsy6m",
      "creatorData": "9999999999aaaaaaaaaabbbbbbbbbbccccccccccddddddddddeeeeeeeeeeffff",
      "salt": "aabbccddeeffffeeddbbccddaa66551155aaaabbcc787878789900aabbccddeeffffeeddbbccddaa66551155aaaabbcc787878789900aabbbbcc221100acadae",
      "msg": "{}"
    },
    "intermediate": {
      "key": "7761736d00000000000000002013a1fc994cc6d1c81b746ee0c0ff6f90043875e0bf1d9be6b7d779fc978dc2a500000000000000209999999999aaaaaaaaaabbbbbbbbbbccccccccccddddddddddeeeeeeeeeeffff0000000000000040aabbccddeeffffeeddbbccddaa66551155aaaabbcc787878789900aabbccddeeffffeeddbbccddaa66551155aaaabbcc787878789900aabbbbcc221100acadae00000000000000027b7d",
      "addressData": "36fe6ab732187cdfed46290b448b32eb7f4798e7a4968b0537de8a842cbf030e"
    },
    "out": {
      "address": "purple1xmlx4dejrp7dlm2x9y95fzejadl50x885jtgkpfhm69ggt9lqv8qk3vn4f"
    }
  },
  {
    "in": {
      "checksum": "13a1fc994cc6d1c81b746ee0c0ff6f90043875e0bf1d9be6b7d779fc978dc2a5",
      "creator": "purple1nxvenxve42424242hwamhwamenxvenxvmhwamhwaamhwamhwlllsatsy6m",
      "creatorData": "9999999999aaaaaaaaaabbbbbbbbbbccccccccccddddddddddeeeeeeeeeeffff",
      "salt": "aabbccddeeffffeeddbbccddaa66551155aaaabbcc787878789900aabbccddeeffffeeddbbccddaa66551155aaaabbcc787878789900aabbbbcc221100acadae",
      "msg": "{\"some\":123,\"structure\":{\"nested\":[\"ok\",true]}}"
    },
    "intermediate": {
      "key": "7761736d00000000000000002013a1fc994cc6d1c81b746ee0c0ff6f90043875e0bf1d9be6b7d779fc978dc2a500000000000000209999999999aaaaaaaaaabbbbbbbbbbccccccccccddddddddddeeeeeeeeeeffff0000000000000040aabbccddeeffffeeddbbccddaa66551155aaaabbcc787878789900aabbccddeeffffeeddbbccddaa66551155aaaabbcc787878789900aabbbbcc221100acadae000000000000002f7b22736f6d65223a3132332c22737472756374757265223a7b226e6573746564223a5b226f6b222c747275655d7d7d",
      "addressData": "a0d0c942adac6f3e5e7c23116c4c42a24e96e0ab75f53690ec2d3de16067c751"
    },
    "out": {
      "address": "purple15rgvjs4d43hnuhnuyvgkcnzz5f8fdc9twh6ndy8v9577zcr8cags40l9dt"
    }
  },
  {
    "in": {
      "checksum": "1da6c16de2cbaf7ad8cbb66f0925ba33f5c278cb2491762d04658c1480ea229b",
      "creator": "purple1nxvenxve42424242hwamhwamenxvenxvhxf2py",
      "creatorData": "9999999999aaaaaaaaaabbbbbbbbbbcccccccccc",
      "salt": "61",
      "msg": null
    },
    "intermediate": {
      "key": "7761736d0000000000000000201da6c16de2cbaf7ad8cbb66f0925ba33f5c278cb2491762d04658c1480ea229b00000000000000149999999999aaaaaaaaaabbbbbbbbbbcccccccccc0000000000000001610000000000000000",
      "addressData": "b95c467218d408a0f93046f227b6ece7fe18133ff30113db4d2a7becdfeca141"
    },
    "out": {
      "address": "purple1h9wyvusc6sy2p7fsgmez0dhvullpsyel7vq38k6d9fa7ehlv59qsvnyh36"
    }
  },
  {
    "in": {
      "checksum": "1da6c16de2cbaf7ad8cbb66f0925ba33f5c278cb2491762d04658c1480ea229b",
      "creator": "purple1nxvenxve42424242hwamhwamenxvenxvhxf2py",
      "creatorData": "9999999999aaaaaaaaaabbbbbbbbbbcccccccccc",
      "salt": "61",
      "msg": "{}"
    },
    "intermediate": {
      "key": "7761736d0000000000000000201da6c16de2cbaf7ad8cbb66f0925ba33f5c278cb2491762d04658c1480ea229b00000000000000149999999999aaaaaaaaaabbbbbbbbbbcccccccccc00000000000000016100000000000000027b7d",
      "addressData": "23fe45dbbd45dc6cd25244a74b6e99e7a65bf0bac2f2842a05049d37555a3ae6"
    },
    "out": {
      "address": "purple1y0lytkaaghwxe5jjgjn5km5eu7n9hu96ctegg2s9qjwnw4268tnqxhg60a"
    }
  },
  {
    "in": {
      "checksum": "1da6c16de2cbaf7ad8cbb66f0925ba33f5c278cb2491762d04658c1480ea229b",
      "creator": "purple1nxvenxve42424242hwamhwamenxvenxvhxf2py",
      "creatorData": "9999999999aaaaaaaaaabbbbbbbbbbcccccccccc",
      "salt": "61",
      "msg": "{\"some\":123,\"structure\":{\"nested\":[\"ok\",true]}}"
    },
    "intermediate": {
      "key": "7761736d0000000000000000201da6c16de2cbaf7ad8cbb66f0925ba33f5c278cb2491762d04658c1480ea229b00000000000000149999999999aaaaaaaaaabbbbbbbbbbcccccccccc000000000000000161000000000000002f7b22736f6d65223a3132332c22737472756374757265223a7b226e6573746564223a5b226f6b222c747275655d7d7d",
      "addressData": "6faea261ed63baa65b05726269e83b217fa6205dc7d9fb74f9667d004a69c082"
    },
    "out": {
      "address": "purple1d7h2yc0dvwa2vkc9wf3xn6pmy9l6vgzaclvlka8eve7sqjnfczpqqsdnwu"
    }
  },
  {
    "in": {
      "checksum": "1da6c16de2cbaf7ad8cbb66f0925ba33f5c278cb2491762d04658c1480ea229b",
      "creator": "purple1nxvenxve42424242hwamhwamenxvenxvhxf2py",
      "creatorData": "9999999999aaaaaaaaaabbbbbbbbbbcccccccccc",
      "salt": "aabbccddeeffffeeddbbccddaa66551155aaaabbcc787878789900aabbccddeeffffeeddbbccddaa66551155aaaabbcc787878789900aabbbbcc221100acadae",
      "msg": null
    },
    "intermediate": {
      "key": "7761736d0000000000000000201da6c16de2cbaf7ad8cbb66f0925ba33f5c278cb2491762d04658c1480ea229b00000000000000149999999999aaaaaaaaaabbbbbbbbbbcccccccccc0000000000000040aabbccddeeffffeeddbbccddaa66551155aaaabbcc787878789900aabbccddeeffffeeddbbccddaa66551155aaaabbcc787878789900aabbbbcc221100acadae0000000000000000",
      "addressData": "67a3ab6384729925fdb144574628ce96836fe098d2c6be4e84ac106b2728d96c"
    },
    "out": {
      "address": "purple1v736kcuyw2vjtld3g3t5v2xwj6pklcyc6trtun5y4sgxkfegm9kq7vgpnt"
    }
  },
  {
    "in": {
      "checksum": "1da6c16de2cbaf7ad8cbb66f0925ba33f5c278cb2491762d04658c1480ea229b",
      "creator": "purple1nxvenxve42424242hwamhwamenxvenxvhxf2py",
      "creatorData": "9999999999aaaaaaaaaabbbbbbbbbbcccccccccc",
      "salt": "aabbccddeeffffeeddbbccddaa66551155aaaabbcc787878789900aabbccddeeffffeeddbbccddaa66551155aaaabbcc787878789900aabbbbcc221100acadae",
      "msg": "{}"
    },
    "intermediate": {
      "key": "7761736d0000000000000000201da6c16de2cbaf7ad8cbb66f0925ba33f5c278cb2491762d04658c1480ea229b00000000000000149999999999aaaaaaaaaabbbbbbbbbbcccccccccc0000000000000040aabbccddeeffffeeddbbccddaa66551155aaaabbcc787878789900aabbccddeeffffeeddbbccddaa66551155aaaabbcc787878789900aabbbbcc221100acadae00000000000000027b7d",
      "addressData": "23a121263bfce05c144f4af86f3d8a9f87dc56f9dc48dbcffc8c5a614da4c661"
    },
    "out": {
      "address": "purple1ywsjzf3mlns9c9z0ftux70v2n7rac4hem3ydhnlu33dxzndycesssc7x2m"
    }
  },
  {
    "in": {
      "checksum": "1da6c16de2cbaf7ad8cbb66f0925ba33f5c278cb2491762d04658c1480ea229b",
      "creator": "purple1nxvenxve42424242hwamhwamenxvenxvhxf2py",
      "creatorData": "9999999999aaaaaaaaaabbbbbbbbbbcccccccccc",
      "salt": "aabbccddeeffffeeddbbccddaa66551155aaaabbcc787878789900aabbccddeeffffeeddbbccddaa66551155aaaabbcc787878789900aabbbbcc221100acadae",
      "msg": "{\"some\":123,\"structure\":{\"nested\":[\"ok\",true]}}"
    },
    "intermediate": {
      "key": "7761736d0000000000000000201da6c16de2cbaf7ad8cbb66f0925ba33f5c278cb2491762d04658c1480ea229b00000000000000149999999999aaaaaaaaaabbbbbbbbbbcccccccccc0000000000000040aabbccddeeffffeeddbbccddaa66551155aaaabbcc787878789900aabbccddeeffffeeddbbccddaa66551155aaaabbcc787878789900aabbbbcc221100acadae000000000000002f7b22736f6d65223a3132332c22737472756374757265223a7b226e6573746564223a5b226f6b222c747275655d7d7d",
      "addressData": "dd90dba6d6fcd5fb9c9c8f536314eb1bb29cb5aa084b633c5806b926a5636b58"
    },
    "out": {
      "address": "purple1mkgdhfkkln2lh8yu3afkx98trwefedd2pp9kx0zcq6ujdftrddvq50esay"
    }
  },
  {
    "in": {
      "checksum": "1da6c16de2cbaf7ad8cbb66f0925ba33f5c278cb2491762d04658c1480ea229b",
      "creator": "purple1nxvenxve42424242hwamhwamenxvenxvmhwamhwaamhwamhwlllsatsy6m",
      "creatorData": "9999999999aaaaaaaaaabbbbbbbbbbccccccccccddddddddddeeeeeeeeeeffff",
      "salt": "61",
      "msg": null
    },
    "intermediate": {
      "key": "7761736d0000000000000000201da6c16de2cbaf7ad8cbb66f0925ba33f5c278cb2491762d04658c1480ea229b00000000000000209999999999aaaaaaaaaabbbbbbbbbbccccccccccddddddddddeeeeeeeeeeffff0000000000000001610000000000000000",
      "addressData": "547a743022f4f1af05b102f57bf1c1c7d7ee81bae427dc20d36b2c4ec57612ae"
    },
    "out": {
      "address": "purple123a8gvpz7nc67pd3qt6hhuwpclt7aqd6usnacgxndvkya3tkz2hq5hz38f"
    }
  },
  {
    "in": {
      "checksum": "1da6c16de2cbaf7ad8cbb66f0925ba33f5c278cb2491762d04658c1480ea229b",
      "creator": "purple1nxvenxve42424242hwamhwamenxvenxvmhwamhwaamhwamhwlllsatsy6m",
      "creatorData": "9999999999aaaaaaaaaabbbbbbbbbbccccccccccddddddddddeeeeeeeeeeffff",
      "salt": "61",
      "msg": "{}"
    },
    "intermediate": {
      "key": "7761736d0000000000000000201da6c16de2cbaf7ad8cbb66f0925ba33f5c278cb2491762d04658c1480ea229b00000000000000209999999999aaaaaaaaaabbbbbbbbbbccccccccccddddddddddeeeeeeeeeeffff00000000000000016100000000000000027b7d",
      "addressData": "416e169110e4b411bc53162d7503b2bbf14d6b36b1413a4f4c9af622696e9665"
    },
    "out": {
      "address": "purple1g9hpdygsuj6pr0znzckh2qajh0c566ekk9qn5n6vntmzy6twjejsrl9alk"
    }
  },
  {
    "in": {
      "checksum": "1da6c16de2cbaf7ad8cbb66f0925ba33f5c278cb2491762d04658c1480ea229b",
      "creator": "purple1nxvenxve42424242hwamhwamenxvenxvmhwamhwaamhwamhwlllsatsy6m",
      "creatorData": "9999999999aaaaaaaaaabbbbbbbbbbccccccccccddddddddddeeeeeeeeeeffff",
      "salt": "61",
      "msg": "{\"some\":123,\"structure\":{\"nested\":[\"ok\",true]}}"
    },
    "intermediate": {
      "key": "7761736d0000000000000000201da6c16de2cbaf7ad8cbb66f0925ba33f5c278cb2491762d04658c1480ea229b00000000000000209999999999aaaaaaaaaabbbbbbbbbbccccccccccddddddddddeeeeeeeeeeffff000000000000000161000000000000002f7b22736f6d65223a3132332c22737472756374757265223a7b226e6573746564223a5b226f6b222c747275655d7d7d",
      "addressData": "619a0988b92d8796cea91dea63cbb1f1aefa4a6b6ee5c5d1e937007252697220"
    },
    "out": {
      "address": "purple1vxdqnz9e9kredn4frh4x8ja37xh05jntdmjut50fxuq8y5nfwgsquu9mxh"
    }
  },
  {
    "in": {
      "checksum": "1da6c16de2cbaf7ad8cbb66f0925ba33f5c278cb2491762d04658c1480ea229b",
      "creator": "purple1nxvenxve42424242hwamhwamenxvenxvmhwamhwaamhwamhwlllsatsy6m",
      "creatorData": "9999999999aaaaaaaaaabbbbbbbbbbccccccccccddddddddddeeeeeeeeeeffff",
      "salt": "aabbccddeeffffeeddbbccddaa66551155aaaabbcc787878789900aabbccddeeffffeeddbbccddaa66551155aaaabbcc787878789900aabbbbcc221100acadae",
      "msg": null
    },
    "intermediate": {
      "key": "7761736d0000000000000000201da6c16de2cbaf7ad8cbb66f0925ba33f5c278cb2491762d04658c1480ea229b00000000000000209999999999aaaaaaaaaabbbbbbbbbbccccccccccddddddddddeeeeeeeeeeffff0000000000000040aabbccddeeffffeeddbbccddaa66551155aaaabbcc787878789900aabbccddeeffffeeddbbccddaa66551155aaaabbcc787878789900aabbbbcc221100acadae0000000000000000",
      "addressData": "d8af856a6a04852d19b647ad6d4336eb26e077f740aef1a0331db34d299a885a"
    },
    "out": {
      "address": "purple1mzhc26n2qjzj6xdkg7kk6sekavnwqalhgzh0rgpnrke562v63pdq8grp8q"
    }
  },
  {
    "in": {
      "checksum": "1da6c16de2cbaf7ad8cbb66f0925ba33f5c278cb2491762d04658c1480ea229b",
      "creator": "purple1nxvenxve42424242hwamhwamenxvenxvmhwamhwaamhwamhwlllsatsy6m",
      "creatorData": "9999999999aaaaaaaaaabbbbbbbbbbccccccccccddddddddddeeeeeeeeeeffff",
      "salt": "aabbccddeeffffeeddbbccddaa66551155aaaabbcc787878789900aabbccddeeffffeeddbbccddaa66551155aaaabbcc787878789900aabbbbcc221100acadae",
      "msg": "{}"
    },
    "intermediate": {
      "key": "7761736d0000000000000000201da6c16de2cbaf7ad8cbb66f0925ba33f5c278cb2491762d04658c1480ea229b00000000000000209999999999aaaaaaaaaabbbbbbbbbbccccccccccddddddddddeeeeeeeeeeffff0000000000000040aabbccddeeffffeeddbbccddaa66551155aaaabbcc787878789900aabbccddeeffffeeddbbccddaa66551155aaaabbcc787878789900aabbbbcc221100acadae00000000000000027b7d",
      "addressData": "c7fb7bea96daab23e416c4fcf328215303005e1d0d5424257335568e5381e33c"
    },
    "out": {
      "address": "purple1clahh65km24j8eqkcn70x2pp2vpsqhsap42zgftnx4tgu5upuv7q9ywjws"
    }
  },
  {
    "in": {
      "checksum": "1da6c16de2cbaf7ad8cbb66f0925ba33f5c278cb2491762d04658c1480ea229b",
      "creator": "purple1nxvenxve42424242hwamhwamenxvenxvmhwamhwaamhwamhwlllsatsy6m",
      "creatorData": "9999999999aaaaaaaaaabbbbbbbbbbccccccccccddddddddddeeeeeeeeeeffff",
      "salt": "aabbccddeeffffeeddbbccddaa66551155aaaabbcc787878789900aabbccddeeffffeeddbbccddaa66551155aaaabbcc787878789900aabbbbcc221100acadae",
      "msg": "{\"some\":123,\"structure\":{\"nested\":[\"ok\",true]}}"
    },
    "intermediate": {
      "key": "7761736d0000000000000000201da6c16de2cbaf7ad8cbb66f0925ba33f5c278cb2491762d04658c1480ea229b00000000000000209999999999aaaaaaaaaabbbbbbbbbbccccccccccddddddddddeeeeeeeeeeffff0000000000000040aabbccddeeffffeeddbbccddaa66551155aaaabbcc787878789900aabbccddeeffffeeddbbccddaa66551155aaaabbcc787878789900aabbbbcc221100acadae000000000000002f7b22736f6d65223a3132332c22737472756374757265223a7b226e6573746564223a5b226f6b222c747275655d7d7d",
      "addressData": "ccdf9dea141a6c2475870529ab38fae9dec30df28e005894fe6578b66133ab4a"
    },
    "out": {
      "address": "purple1en0em6s5rfkzgav8q556kw86a80vxr0j3cq93987v4utvcfn4d9q0tql4w"
    }
  }
]
`
