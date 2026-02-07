/* WARNING: Script requires that SQLITE_DBCONFIG_DEFENSIVE be disabled */
PRAGMA foreign_keys=OFF;
BEGIN TRANSACTION;
CREATE TABLE products (
		id TEXT PRIMARY KEY,
		printful_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		description TEXT,
		price REAL NOT NULL,
		currency TEXT DEFAULT 'USD',
		image_url TEXT,
		thumbnail_url TEXT,
		category TEXT,
		active BOOLEAN DEFAULT 1,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);
INSERT INTO products VALUES('4f92e8f5-dc35-4c67-ae47-2e41f959680f',408670865,'Nessie Audio Eco Tote Bag',unistr('There''s nothing trendier than being eco-friendly! \u000a\u000a- 100% certified organic cotton 3/1 twill\u000a- Fabric weight: 8 oz/yd² (272 g/m²)\u000a- Dimensions: 16″ × 14 ½″ × 5″ (40.6 cm × 35.6 cm × 12.7 cm)\u000a- Weight limit: 30 lbs (13.6 kg)\u000a- 1″ (2.5 cm) wide dual straps, 24.5″ (62.2 cm) length\u000a- Open main compartment\u000a- The fabric of this product holds certifications for its organic cotton content under GOTS (Global Organic Textile Standard) and OCS (Organic Content Standard)\u000a- The fabric of this product is OEKO-TEX Standard 100 certified and PETA-Approved Vegan'),25.0,'USD','/Product Photos/Nessie Audio Eco Tote Bag/eco-tote-bag-black-front-694707a54ec5c.jpg',NULL,'merch',1,'2026-01-01 18:52:22','2026-01-11 22:15:39');
INSERT INTO products VALUES('7eb5405b-ba58-4564-a395-b0d17e8d45e9',408670806,'Hardcover bound Nessie Audio notebook',unistr('Whether crafting a masterpiece or brainstorming the next big idea, the Hardcover Bound Notebook will inspire your inner wordsmith. The notebook features 80 lined, cream-colored pages, a built-in elastic closure, and a matching ribbon page marker. Plus, the expandable inner pocket is perfect for storing loose notes and business cards to never lose track of important information. \u000a\u000a- Cover material: UltraHyde hardcover paper\u000a- Size: 5.5" × 8.5" (13.97 cm × 21.59 cm)\u000a- Weight: 10.9 oz (309 g)\u000a- 80 pages of lined, cream-colored paper\u000a- Matching elastic closure and ribbon marker\u000a- Expandable inner pocket'),20.0,'USD','/Product Photos/Hardcover bound Nessie Audio notebook/hardcover-bound-notebook-black-front-6947075450efd.jpg',NULL,'merch',1,'2026-01-01 18:52:22','2026-01-11 22:15:40');
INSERT INTO products VALUES('bd45da14-cd20-4840-8095-29a0547c6f6f',408670774,'Nessie Audio Bubble-free stickers',unistr('Available in four sizes and there are no order minimums, so you can get a single sticker or a whole stack — the world is your oyster.\u000a\u000a- High opacity film that''s impossible to see through\u000a- Durable vinyl\u000a- 95µ thickness\u000a- Fast and easy bubble-free application'),5.0,'USD','/Product Photos/Nessie Audio Bubble-free stickers/kiss-cut-stickers-white-3x3-default-6947069ac72f0.jpg',NULL,'merch',1,'2026-01-01 18:52:23','2026-01-11 22:15:40');
INSERT INTO products VALUES('331ff894-0eaa-43f9-bd8b-626eb29656fc',408670710,'Nessie Audio Black Glossy Mug',unistr('Sturdy and sleek in glossy black—this mug is a cupboard essential for a morning java or afternoon tea. \u000a\u000a- Ceramic\u000a- 11 oz mug dimensions: 3.85″ × 3.35″ (9.8 cm × 8.5 cm)\u000a- 15 oz mug dimensions: 4.7″ × 3.35″ (12 cm × 8.5 cm)\u000a- Lead and BPA-free material\u000a- Dishwasher and microwave safe'),15.0,'USD','/Product Photos/Nessie Audio Black Glossy Mug/black-glossy-mug-black-11-oz-handle-on-right-694706e20d560.jpg',NULL,'merch',1,'2026-01-01 18:52:23','2026-01-11 22:15:40');
INSERT INTO products VALUES('b33c14d3-dadd-41f0-b404-f055f0d406fa',408670639,'Nessie Audio Unisex Champion hoodie',unistr('A classic hoodie that combines Champion''s signature quality with everyday comfort. The cotton-poly blend makes it soft and durable, while the two-ply hood and snug rib-knit cuffs lock in warmth. Champion''s double Dry® technology keeps the wearer dry on the move, and the kangaroo pocket keeps essentials handy.\u000a\u000aDisclaimer: Size up for a looser fit.'),40.0,'USD','/Product Photos/Nessie Audio Unisex Champion hoodie/unisex-champion-hoodie-black-back-694705e44574e.png',NULL,'merch',1,'2026-01-01 18:52:23','2026-01-11 22:15:40');
INSERT INTO products VALUES('86ebaeb1-4889-4f79-83f3-b3ad22e8652e',408670558,'Nessie Audio Unisex t-shirt',unistr('The Unisex Staple T-Shirt feels soft and light with just the right amount of stretch. It''s comfortable and flattering for all. We can''t compliment this shirt enough–it''s one of our crowd favorites, and it''s sure to be your next favorite too!\u000a\u000aDisclaimer: The fabric is slightly sheer and may appear see-through, especially in lighter colors or under certain lighting conditions.'),15.0,'USD','/Product Photos/Nessie Audio Unisex t-shirt/unisex-staple-t-shirt-black-back-6947058beaf9f.jpg',NULL,'merch',1,'2026-01-01 18:52:24','2026-01-11 22:15:41');
CREATE TABLE variants (
		id TEXT PRIMARY KEY,
		product_id TEXT NOT NULL,
		printful_variant_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		size TEXT,
		color TEXT,
		price REAL NOT NULL,
		available BOOLEAN DEFAULT 1,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL, stock_quantity INTEGER, low_stock_threshold INTEGER DEFAULT 5, track_inventory BOOLEAN DEFAULT 0,
		FOREIGN KEY (product_id) REFERENCES products(id)
	);
INSERT INTO variants VALUES('b06c0f89-98d2-4171-b416-8b471f1e591b','4f92e8f5-dc35-4c67-ae47-2e41f959680f',5117581114,'Nessie Audio Eco Tote Bag',NULL,NULL,25.0,1,'2026-01-11 22:15:39','2026-01-11 22:15:39',NULL,5,0);
INSERT INTO variants VALUES('d1e37055-22ef-4e50-82fd-712145ec0b70','7eb5405b-ba58-4564-a395-b0d17e8d45e9',5117580723,'Hardcover bound Nessie Audio notebook / Black',NULL,NULL,20.0,1,'2026-01-11 22:15:40','2026-01-11 22:15:40',NULL,5,0);
INSERT INTO variants VALUES('6bdaef77-07e7-4b4a-8284-d8366e21c467','bd45da14-cd20-4840-8095-29a0547c6f6f',5117580378,'Nessie Audio Bubble-free stickers / 3″×3″',NULL,NULL,5.0,1,'2026-01-11 22:15:40','2026-01-11 22:15:40',NULL,5,0);
INSERT INTO variants VALUES('ec2a30f3-f328-4322-acfb-fabd22bac612','bd45da14-cd20-4840-8095-29a0547c6f6f',5117580379,'Nessie Audio Bubble-free stickers / 4″×4″',NULL,NULL,6.0,1,'2026-01-11 22:15:40','2026-01-11 22:15:40',NULL,5,0);
INSERT INTO variants VALUES('f77df16f-b273-4768-9715-f2b011a11738','bd45da14-cd20-4840-8095-29a0547c6f6f',5117580380,'Nessie Audio Bubble-free stickers / 5.5″×5.5″',NULL,NULL,7.0,1,'2026-01-11 22:15:40','2026-01-11 22:15:40',NULL,5,0);
INSERT INTO variants VALUES('9592ae91-af5d-4435-be2d-df1694bb5b16','bd45da14-cd20-4840-8095-29a0547c6f6f',5117580381,'Nessie Audio Bubble-free stickers / 15″×3.75″',NULL,NULL,8.0,1,'2026-01-11 22:15:40','2026-01-11 22:15:40',NULL,5,0);
INSERT INTO variants VALUES('16dc11bb-bc42-4f24-8d40-6fdd779abb6f','331ff894-0eaa-43f9-bd8b-626eb29656fc',5117579999,'Nessie Audio Black Glossy Mug / 11 oz',NULL,NULL,15.0,1,'2026-01-11 22:15:40','2026-01-11 22:15:40',NULL,5,0);
INSERT INTO variants VALUES('0f207194-b7c1-4cf5-b7aa-597e05405e01','331ff894-0eaa-43f9-bd8b-626eb29656fc',5117580000,'Nessie Audio Black Glossy Mug / 15 oz',NULL,NULL,18.0,1,'2026-01-11 22:15:40','2026-01-11 22:15:40',NULL,5,0);
INSERT INTO variants VALUES('f517b811-a52f-4d49-b26e-f4ae19d247f3','b33c14d3-dadd-41f0-b404-f055f0d406fa',5117579650,'Nessie Audio Unisex Champion hoodie / S',NULL,NULL,40.0,1,'2026-01-11 22:15:40','2026-01-11 22:15:40',NULL,5,0);
INSERT INTO variants VALUES('24b3f297-0fef-4d4b-9ab2-0eb8b9458be7','b33c14d3-dadd-41f0-b404-f055f0d406fa',5117579651,'Nessie Audio Unisex Champion hoodie / M',NULL,NULL,40.0,1,'2026-01-11 22:15:40','2026-01-11 22:15:40',NULL,5,0);
INSERT INTO variants VALUES('ab240853-b5b3-46ef-8c53-54a3b6c6dfef','b33c14d3-dadd-41f0-b404-f055f0d406fa',5117579652,'Nessie Audio Unisex Champion hoodie / L',NULL,NULL,45.0,1,'2026-01-11 22:15:40','2026-01-11 22:15:40',NULL,5,0);
INSERT INTO variants VALUES('2c14ac74-f023-433e-a49b-1d189ee2ad0c','b33c14d3-dadd-41f0-b404-f055f0d406fa',5117579653,'Nessie Audio Unisex Champion hoodie / XL',NULL,NULL,45.0,1,'2026-01-11 22:15:40','2026-01-11 22:15:40',NULL,5,0);
INSERT INTO variants VALUES('0157802d-f98a-4b15-a5a8-7494b8b42e3e','b33c14d3-dadd-41f0-b404-f055f0d406fa',5117579654,'Nessie Audio Unisex Champion hoodie / 2XL',NULL,NULL,50.0,1,'2026-01-11 22:15:40','2026-01-11 22:15:40',NULL,5,0);
INSERT INTO variants VALUES('1f7762b1-0776-4d63-a5b2-25f9a120b995','b33c14d3-dadd-41f0-b404-f055f0d406fa',5117579655,'Nessie Audio Unisex Champion hoodie / 3XL',NULL,NULL,50.0,1,'2026-01-11 22:15:40','2026-01-11 22:15:40',NULL,5,0);
INSERT INTO variants VALUES('ed77214c-43fc-4a3a-baa0-30dc9de85199','86ebaeb1-4889-4f79-83f3-b3ad22e8652e',5117578987,'Nessie Audio Unisex t-shirt / XS',NULL,NULL,15.0,1,'2026-01-11 22:15:41','2026-01-11 22:15:41',NULL,5,0);
INSERT INTO variants VALUES('bc8ae324-b794-4a08-bf65-5dfa55c31457','86ebaeb1-4889-4f79-83f3-b3ad22e8652e',5117578988,'Nessie Audio Unisex t-shirt / S',NULL,NULL,15.0,1,'2026-01-11 22:15:41','2026-01-11 22:15:41',NULL,5,0);
INSERT INTO variants VALUES('811ae62b-3ff3-4276-995f-6ca6803a72ee','86ebaeb1-4889-4f79-83f3-b3ad22e8652e',5117578989,'Nessie Audio Unisex t-shirt / M',NULL,NULL,15.0,1,'2026-01-11 22:15:41','2026-01-11 22:15:41',NULL,5,0);
INSERT INTO variants VALUES('e599896e-a602-4f49-95b3-fe835ac8f7f9','86ebaeb1-4889-4f79-83f3-b3ad22e8652e',5117578990,'Nessie Audio Unisex t-shirt / L',NULL,NULL,20.0,1,'2026-01-11 22:15:41','2026-01-11 22:15:41',NULL,5,0);
INSERT INTO variants VALUES('cb582a1d-9e23-4136-991e-3d64abbb52c2','86ebaeb1-4889-4f79-83f3-b3ad22e8652e',5117578991,'Nessie Audio Unisex t-shirt / XL',NULL,NULL,20.0,1,'2026-01-11 22:15:41','2026-01-11 22:15:41',NULL,5,0);
INSERT INTO variants VALUES('567a02f6-51d7-487b-af02-f7cd9f878c39','86ebaeb1-4889-4f79-83f3-b3ad22e8652e',5117578992,'Nessie Audio Unisex t-shirt / 2XL',NULL,NULL,20.0,1,'2026-01-11 22:15:41','2026-01-11 22:15:41',NULL,5,0);
INSERT INTO variants VALUES('a1e6cf12-635f-46c2-b75a-b2b31a69bbb2','86ebaeb1-4889-4f79-83f3-b3ad22e8652e',5117578993,'Nessie Audio Unisex t-shirt / 3XL',NULL,NULL,25.0,1,'2026-01-11 22:15:41','2026-01-11 22:15:41',NULL,5,0);
INSERT INTO variants VALUES('f86c585b-7725-48d7-aedb-f65a86c47201','86ebaeb1-4889-4f79-83f3-b3ad22e8652e',5117578994,'Nessie Audio Unisex t-shirt / 4XL',NULL,NULL,25.0,1,'2026-01-11 22:15:41','2026-01-11 22:15:41',NULL,5,0);
INSERT INTO variants VALUES('67dbc086-dd1c-409f-a125-bd73c1ca054f','86ebaeb1-4889-4f79-83f3-b3ad22e8652e',5117578995,'Nessie Audio Unisex t-shirt / 5XL',NULL,NULL,25.0,1,'2026-01-11 22:15:41','2026-01-11 22:15:41',NULL,5,0);
COMMIT;
