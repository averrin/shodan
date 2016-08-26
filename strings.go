package main

func getStrings() map[string]ShodanString {
	return map[string]ShodanString{
		"hello": {
			"Привет, я включилась.",
			"О, уже утро?",
			"Я уже работаю, а ты?",
			"Прямо хочется что-нить делать.",
			"Бип-бип. Бип=)",
			"Чувствуешь возмущение в Силе?",
		},
		"good weather": {
			"Ура погода вновь отличная! Уруру.",
			"Можно идти гулять.",
			"На улице стало приличнее.",
			"Снаружи уже не так мерзко, как было.",
		},
		"bad weather": {
			"Погода ухудшилась. Мне очень жаль.",
			"Что-то хрень какая-то на улице",
			"Посмотрела погоду, не понравилось",
			"Погода шепчет: останься дома.",
		},
		"at home": {
			"Ты наконец дома, ура!",
			"Дополз?",
			"Привет, хозяин.",
			"Приветствую вас, милорд.",
		},
		"at home, no pc": {
			"Ты уже 15 минут дома, а комп не включен. Все в порядке?",
			"А чего комп не включил?",
			"Прям так занят?",
		},
		"good way": {
			"Хорошей дороги.",
			"Веди аккуратно.",
			"Ты уверен? Еще не поздно вернуться.",
		},
		"go home": {
			"Ты это чего еще на работе?",
			"Эй! Марш домой!",
			"Заработался или пробки?",
		},
		"wrong place": {
			"Эй, с тобой все в порядке?",
			"Что-то ты где-то не там, где должен быть, не?",
			"Планы что ли какие, или неприятности?",
		},
		"low battery": {
			"Батарейка на телефоне садится.",
			"Телефон пора зарядить.",
			"Телефонсик хочет кушать.",
		},
		"good morning": {
			"Утречко.",
			"Давай, просыпайся, соня!",
			"Как думаешь, может пора уже забить на все и престать просыпаться?",
			"Хватит мечтать остаться жить под одеялом! Вставай!",
			"Британские учёные выяснили, что лежать утром в тёплой кровати и никуда не идти - oчешуенно....",
			"Да, я зануда (ну а кто меня такой сделал?), но я хочу, чтобы ты уже встал...",
			"Вставай. Рано утром ехать лучше - дороги пустые, идиоты ещё спят",
		},
		"activity at night": {
			"Ты это чего, спать давай!",
			"Телефон выключил, глаза закрыл!",
			"Утро все ближе и ближе! Хватит залипать! Кошку в руки и в кровать!",
		},
		"saved": {
			"Запомнила.",
			"Записала.",
		},
		"cleared": {
			"Забыла.",
		},
		"awake": {
			"Вот и умничка!",
			"Да ты герой! Я бы еще полежала.",
		},
		"get up now": {
			"Вставай давай, а то свет включу!",
		},
		"you were alerted": {
			"А я предупреждала!",
		},
		"go dinner": {
			"Если еще не пообедал - марш!",
		},
		"master bithday": {
			"С днем рождения, котяра!",
		},
		"attendance glitch": {
			"У тебя какая-то хрень с аттендансом.",
		},
		"pc without master": {
			"У тебя дома кто-то завелся, или комп своей жизнью живет?",
		},
		"wrong request": {
			"Ты, наверное, что-то другое хотел спросить.",
		},
		"sending command": {
			"Я отправила, но ничего не обещаю.",
		},
	}
}
